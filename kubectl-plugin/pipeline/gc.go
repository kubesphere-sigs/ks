package pipeline

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/go-openapi/strfmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/option"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	devopsclient "github.com/kubesphere/ks-devops-client-go/client"
	"github.com/kubesphere/ks-devops-client-go/client/dev_ops_pipeline"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func newGCCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &gcOption{
		client: client,
		PipelineCreateOption: option.PipelineCreateOption{
			Client: client,
		},
	}
	cmd = &cobra.Command{
		Use:     "gc",
		Short:   "Garbage collector for PipelineRuns",
		Long:    "Clean all those old PipelineRuns by age or count",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opt.cleanPipelinerun, "clean-pipelinerun", "", false,
		"if delete outdated pipelineruns of DevOps project, default: false means only gc pipelineruns by pipeline.discarder")
	flags.UintVarP(&opt.maxCount, "max-count", "", 30,
		"Maximum number to keep PipelineRuns of DevOps project(works when clean-pipelinerun is true)")
	flags.DurationVarP(&opt.maxAge, "max-age", "", 7*24*time.Hour,
		"Maximum age to keep PipelineRuns of DevOps project(works when clean-pipelinerun is true)")
	flags.StringVarP(&opt.condition, "condition", "", conditionAnd,
		fmt.Sprintf("The condition between --max-count and --max-age, supported conditions: '%s', '%s'", conditionAnd, conditionIgnore))
	flags.StringArrayVarP(&opt.namespaces, "namespaces", "", nil,
		"Indicate namespaces do you want to clean. Clean all namespaces if it's empty")
	flags.BoolVarP(&opt.abortPipelinerun, "abort-pipelinerun", "", false,
		"Whether abort pipelineruns that does not finished")
	flags.DurationVarP(&opt.ageToAbort, "age-to-abort", "", 7*24*time.Hour,
		"If a pipelinerun has been created than this age and has not finished yet, it will be aborted")
	flags.StringVarP(&opt.devopsAPIHost, "devops-api-host", "", "devops-apiserver.kubesphere-devops-system.svc:9090",
		"The devops apiserver address")
	flags.StringArrayVarP(&opt.devopsAPISchemes, "devops-api-schemes", "", []string{"http"},
		"The schemes to connect to devops apiserver")
	_ = cmd.RegisterFlagCompletionFunc("condition", common.ArrayCompletion(conditionAnd, conditionIgnore))
	return
}

const (
	conditionAnd    = "and"
	conditionIgnore = "ignoreTime"
)

type gcOption struct {
	cleanPipelinerun bool
	maxCount         uint
	maxAge           time.Duration
	condition        string
	namespaces       []string
	abortPipelinerun bool
	ageToAbort       time.Duration
	devopsAPIHost    string
	devopsAPISchemes []string

	// inner fields
	client dynamic.Interface
	option.PipelineCreateOption
	devopsClient *devopsclient.KubeSphereDevOps
}

func (o *gcOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	if len(o.namespaces) == 0 {
		if err = o.getAllDevOpsNamespace(); err != nil {
			log.Errorf("failed to get all DevOps project namespace, error: %+v", err)
			return
		}
	}
	err = o.initDevopsClient()
	return
}

func (o *gcOption) initDevopsClient() error {
	cfg := &devopsclient.TransportConfig{
		Host:    o.devopsAPIHost,
		Schemes: o.devopsAPISchemes,
	}
	o.devopsClient = devopsclient.NewHTTPClientWithConfig(strfmt.Default, cfg)
	return nil
}

func (o *gcOption) getAllDevOpsNamespace() (err error) {
	var wsList *unstructured.UnstructuredList
	if wsList, err = o.client.Resource(types.GetNamespaceSchema()).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "kubesphere.io/devopsproject",
	}); err == nil {
		o.namespaces = make([]string, len(wsList.Items))
		for i, item := range wsList.Items {
			o.namespaces[i] = item.GetName()
		}
	}
	return
}

func (o *gcOption) cleanPipelineRunInNamespace(namespace string) (err error) {
	var pipelinerunList *unstructured.UnstructuredList
	if pipelinerunList, err = o.GetUnstructuredListInNamespace(namespace, types.GetPipelineRunSchema()); err != nil {
		err = fmt.Errorf("failed to get PipelineRun list, error: %v", err)
		return
	}

	items := pipelinerunList.Items
	toDelete := len(items) - int(o.maxCount)
	if toDelete < 1 {
		return
	}

	ascOrderWithCompletionTime(pipelinerunList.Items)

	for i := range items {
		item := items[i]

		// check remain amount
		if toDelete <= 0 {
			break
		}

		if (o.condition == conditionAnd && okToDelete(item.Object, o.maxAge)) || o.condition == conditionIgnore {
			delErr := o.client.Resource(types.GetPipelineRunSchema()).Namespace(namespace).Delete(
				context.TODO(), item.GetName(), metav1.DeleteOptions{})
			if delErr != nil {
				log.Errorf("failed to delete PipelineRun %s/%s, error: %v", item.GetName(), namespace, delErr)
			} else {
				toDelete--
				log.Errorf("ok to delete PipelineRun %s/%s", item.GetName(), namespace)
			}
		}
	}
	return
}

func (o *gcOption) runE(cmd *cobra.Command, args []string) error {
	// clean pipelinerun of pipeline with days_to_keep and num_to_keep
	for _, ns := range o.namespaces {
		unList, err := o.GetUnstructuredListInNamespace(ns, types.GetPipelineSchema())
		if err != nil {
			cmd.PrintErrf("failed to get Pipeline in '%s', error: %+v\n", ns, err)
			return err
		}

		for _, un := range unList.Items {
			log.Infof("### found pipeline: %s in namespace: %s", un.GetName(), ns)
			pipeline, err := toPipeline(o, un)
			if err != nil {
				cmd.PrintErrf("parse unstructured pipeline to gcPipeline(%s) failed, err: %+v\n", un.GetName(), err)
				continue
			}
			if err := pipeline.getPipelinerun(); err != nil {
				cmd.PrintErrf("failed to get pipelinerun, error: %+v\n", err)
				continue
			}
			if err := pipeline.clean(); err != nil {
				cmd.PrintErrf("clean pipelinerun error: %+v\n", err)
				continue
			}
			if err := pipeline.abort(); err != nil {
				cmd.PrintErrf("abort pipelinerun error: %+v\n", err)
				continue
			}
		}
	}

	if !o.cleanPipelinerun {
		return nil
	}
	log.Info("clean pipelinerun of dev-project by max-count and max-age ..")

	// keep below log output until replace it with a logger
	//cmd.Printf("starting to gc PipelineRuns in %d namespaces\n", len(o.namespaces))
	errorsNs := []string{}

	for i := range o.namespaces {
		ns := o.namespaces[i]
		if err := o.cleanPipelineRunInNamespace(ns); err != nil {
			cmd.PrintErrf("failed to clean PipelineRuns in '%s'\n", ns)
			errorsNs = append(errorsNs, ns)
		}
	}

	if len(errorsNs) > 0 {
		log.Errorf("gc failed in %d namespaces: %v", len(errorsNs), errorsNs)
		return fmt.Errorf("gc failed")
	}
	return nil
}

type gcPipeline struct {
	option *gcOption

	name      string
	namespace string
	pType     string

	discard    bool
	daysToKeep int
	numToKeep  int

	pipelinerunList []*gcPipelinerun
}

func (p *gcPipeline) getPipelinerun() (err error) {
	ctx := context.TODO()
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", option.PipelinerunOwnerLabelKey, p.name),
	}
	var wsList *unstructured.UnstructuredList
	if wsList, err = p.option.Client.Resource(types.GetPipelineRunSchema()).Namespace(p.namespace).List(ctx, opts); err != nil {
		return err
	}

	var pr *gcPipelinerun
	for _, item := range wsList.Items {
		if pr, err = toPipelinerun(item); err != nil {
			return err
		}
		p.pipelinerunList = append(p.pipelinerunList, pr)
	}
	return nil
}

func (p *gcPipeline) ascPipelinerun() {
	if p.pipelinerunList != nil {
		sort.Slice(p.pipelinerunList, func(i, j int) bool {
			leftTime := p.pipelinerunList[i].completionTime
			rightTime := p.pipelinerunList[j].completionTime

			if leftTime.IsZero() {
				return false
			}
			if rightTime.IsZero() {
				// make sure that item without completion time be at the end of items
				return true
			}

			return leftTime.Before(rightTime)
		})
	}
}

func (p *gcPipeline) clean() (err error) {
	log.Infof("clean pipelinerun of pipeline: %s ..", p.name)
	if !p.discard {
		log.Warn("the discarder of pipeline not found, ignore.")
		return
	}
	if p.pType != option.NoScmPipelineType {
		log.Warnf("the type of pipeline is %s, ignore.", p.pType)
		return
	}

	if len(p.pipelinerunList) == 0 {
		log.Infof("there is no pipelinerun of pipeline: %s.", p.name)
		return nil
	}

	deletingPipelinerunList := p.needToDelete()
	for _, run := range deletingPipelinerunList {
		log.Infof("delete pipelinerun: %s/%s ...", run.id, run.name)
		if err = p.option.client.Resource(types.GetPipelineRunSchema()).Namespace(p.namespace).Delete(
			context.TODO(), run.name, metav1.DeleteOptions{}); err != nil {
			log.Errorf("failed to delete PipelineRun: %s, error: %+v", run.name, err)
			return err
		}
		log.Infof("pipelinerun: %s deleted.", run.name)
	}
	return
}

func (p *gcPipeline) abort() (err error) {
	ctx := context.Background()
	if !p.option.abortPipelinerun {
		log.Infof("the abortPipelinerun flag is not enabled")
		return nil
	}
	log.Infof("abort pipelinerun of pipeline: %s ..", p.name)
	for _, run := range p.pipelinerunList {
		if !run.isCompletion() && run.creationTime.Add(p.option.ageToAbort).Before(time.Now()) {
			log.Infof("abort pipelinerun: %s ..", run.name)
			abortErr := p.abortPipelinerun(ctx, run)
			if abortErr != nil {
				log.Errorf("failed to abort PipelineRun: %s, error: %+v", run.name, abortErr)
				// we want to try all pipelineruns, so continue here
				continue
			}
		}
	}
	return nil
}

func (p *gcPipeline) abortPipelinerun(ctx context.Context, run *gcPipelinerun) error {
	if run.pType == option.MultiBranchPipelineType {
		return p.abortMultiBranchPipelinerun(ctx, run)
	}
	return p.abortNoScmPipelinerun(ctx, run)
}

func (p *gcPipeline) abortNoScmPipelinerun(ctx context.Context, run *gcPipelinerun) error {
	stopPipelineParams := &dev_ops_pipeline.StopPipelineParams{
		Blocking:      aws.String("true"),
		Body:          []int64{},
		Devops:        p.namespace,
		Pipeline:      p.name,
		Run:           run.id,
		TimeOutInSecs: aws.String("10"),
		Context:       ctx,
	}
	_, err := p.option.devopsClient.DevOpsPipeline.StopPipeline(stopPipelineParams)
	return err
}

func (p *gcPipeline) abortMultiBranchPipelinerun(ctx context.Context, run *gcPipelinerun) error {
	stopPipelineParams := &dev_ops_pipeline.StopBranchPipelineParams{
		Blocking:      aws.String("true"),
		Body:          []int64{},
		Branch:        run.branch,
		Devops:        p.namespace,
		Pipeline:      p.name,
		Run:           run.id,
		TimeOutInSecs: aws.String("10"),
		Context:       ctx,
	}
	_, err := p.option.devopsClient.DevOpsPipeline.StopBranchPipeline(stopPipelineParams)
	return err
}

func (p *gcPipeline) needToDelete() (deleting []*gcPipelinerun) {
	p.ascPipelinerun()

	// get index of last_successful and last_stable pipelinerun
	var lastSuccessfulIndex, lastStableIndex int
	lastSuccessfulIndex = -1
	lastStableIndex = -1
	for i, pipelinerun := range p.pipelinerunList {
		if pipelinerun.isCompletion() {
			if pipelinerun.phase == option.PipelinerunPhaseSucceeded {
				lastSuccessfulIndex = i
			}
			lastStableIndex = i
		}
	}

	// clean by num_to_keep and day_to_keep
	durationToKeep := time.Duration(p.daysToKeep*24) * time.Hour
	numLimitIndex := len(p.pipelinerunList) - p.numToKeep
	for i, pipelinerun := range p.pipelinerunList {
		if pipelinerun.isCompletion() {
			if i < numLimitIndex {
				if i == lastSuccessfulIndex || i == lastStableIndex { // ignore to delete last-stable and last-successful pipelinerun
					numLimitIndex = numLimitIndex + 1
				} else {
					deleting = append(deleting, pipelinerun)
				}
			} else if pipelinerun.isOverdue(durationToKeep) {
				deleting = append(deleting, pipelinerun)
			}
		}
	}
	return
}

type gcPipelinerun struct {
	id             string
	name           string
	phase          string
	pType          string
	branch         string
	completionTime time.Time
	creationTime   time.Time
}

func (r *gcPipelinerun) isOverdue(maxAge time.Duration) bool {
	return r.completionTime.Add(maxAge).Before(time.Now())
}

func (r *gcPipelinerun) isCompletion() bool {
	return !r.completionTime.IsZero()
}

func toPipeline(gcOpt *gcOption, u unstructured.Unstructured) (*gcPipeline, error) {
	pipeline := &gcPipeline{
		option:    gcOpt,
		name:      u.GetName(),
		namespace: u.GetNamespace(),
		discard:   false,
	}

	t, ok, err := unstructured.NestedString(u.Object, "spec", "type")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("field type not found of pipeline: %s", u.GetName())
	}
	pipeline.pType = t
	contentKey := t
	if t == option.MultiBranchPipelineType {
		contentKey = "multi_branch_pipeline"
	}

	if _, ok, err = unstructured.NestedMap(u.Object, "spec", contentKey, "discarder"); err == nil {
		if ok {
			pipeline.discard = true
			var days, num string
			days, ok, err = unstructured.NestedString(u.Object, "spec", contentKey, "discarder", "days_to_keep")
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("field days_to_keep not found of pipeline: %s", u.GetName())
			}
			if pipeline.daysToKeep, err = strconv.Atoi(days); err != nil {
				return nil, err
			}

			num, ok, err = unstructured.NestedString(u.Object, "spec", contentKey, "discarder", "num_to_keep")
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("field days_to_keep not found of pipeline: %s", u.GetName())
			}
			if pipeline.numToKeep, err = strconv.Atoi(num); err != nil {
				return nil, err
			}
		}
	}
	return pipeline, err
}

func toPipelinerun(u unstructured.Unstructured) (*gcPipelinerun, error) {
	creationTime := u.GetCreationTimestamp().Time

	pType, ok, err := unstructured.NestedString(u.Object, "spec", "pipelineSpec", "type")
	if err != nil {
		return nil, err
	}
	if !ok {
		pType = option.NoScmPipelineType
	}

	var branch string
	if pType == option.MultiBranchPipelineType {
		branch, ok, err = unstructured.NestedString(u.Object, "spec", "scm", "refName")
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("the spec.scm.refName of pipelinerun: %s not found", u.GetName())
		}
	}

	phase, ok, err := unstructured.NestedString(u.Object, "status", "phase")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("the phase of pipelinerun: %s not found", u.GetName())
	}
	id := u.GetAnnotations()[option.PipelinerunIdAnnotationKey]

	pipelinerun := &gcPipelinerun{
		id:           id,
		name:         u.GetName(),
		phase:        phase,
		pType:        pType,
		branch:       branch,
		creationTime: creationTime,
	}
	if phase == option.PipelinerunPhaseSucceeded || phase == option.PipelinerunPhaseFailed || phase == option.PipelinerunPhaseCancelled {
		pipelinerun.completionTime, err = getCompletionTimeFromObject(u.Object)
	}
	return pipelinerun, err
}

func ascOrderWithCompletionTime(items []unstructured.Unstructured) {
	sort.Slice(items, func(i, j int) bool {
		left := items[i]
		right := items[j]

		var leftCompletionTime time.Time
		var rightCompletionTime time.Time
		var err error

		if leftCompletionTime, err = getCompletionTimeFromObject(left.Object); err != nil {
			return false
		}
		if rightCompletionTime, err = getCompletionTimeFromObject(right.Object); err != nil {
			// make sure that item without completion time be at the end of items
			return true
		}

		return leftCompletionTime.Before(rightCompletionTime)
	})
}

func getCompletionTimeFromObject(obj map[string]interface{}) (completionTime time.Time, err error) {
	var (
		completionTimeStr string
		ok                bool
	)
	if completionTimeStr, ok, err = unstructured.NestedString(obj, "status", "completionTime"); ok && err == nil {
		completionTime, err = time.Parse(time.RFC3339, completionTimeStr)
	}
	if !ok {
		err = errors.New("no status.completionTime field found")
	}
	return
}

func okToDelete(object map[string]interface{}, maxAge time.Duration) bool {
	if completionTime, err := getCompletionTimeFromObject(object); err == nil {
		return completionTime.Add(maxAge).Before(time.Now())
	}
	return false
}
