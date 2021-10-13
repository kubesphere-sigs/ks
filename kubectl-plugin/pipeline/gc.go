package pipeline

import (
	"context"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/option"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"sort"
	"time"
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
	flags.IntVarP(&opt.maxCount, "max-count", "", 30,
		"Maximum number to keep PipelineRuns")
	flags.DurationVarP(&opt.maxAge, "max-age", "", 7*24*time.Hour,
		"Maximum age to keep PipelineRuns")
	flags.StringVarP(&opt.condition, "condition", "", conditionAnd,
		fmt.Sprintf("The condition between --max-count and --max-age, supported conditions: '%s', '%s'", conditionAnd, conditionIgnore))
	flags.StringArrayVarP(&opt.namespaces, "namespaces", "", nil,
		"Indicate namespaces do you want to clean. Clean all namespaces if it's empty")

	_ = cmd.RegisterFlagCompletionFunc("condition", common.ArrayCompletion(conditionAnd, conditionIgnore))
	return
}

const (
	conditionAnd    = "and"
	conditionIgnore = "ignoreTime"
)

type gcOption struct {
	maxCount   int
	maxAge     time.Duration
	condition  string
	namespaces []string

	// inner fields
	client dynamic.Interface
	option.PipelineCreateOption
}

func (o *gcOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	if len(o.namespaces) == 0 {
		o.namespaces = getAllNamespace(o.client)
	}
	return
}

func (o *gcOption) cleanPipelineRunInNamespace(namespace string) (err error) {
	var pipelineList *unstructured.UnstructuredList
	if pipelineList, err = o.GetUnstructuredListInNamespace(namespace, types.GetPipelineRunSchema()); err != nil {
		err = fmt.Errorf("failed to get PipelineRun list, error: %v", err)
		return
	}

	items := pipelineList.Items
	toDelete := len(items) - o.maxCount
	if toDelete < o.maxCount {
		return
	}

	ascOrderWithCompletionTime(pipelineList.Items)

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
				fmt.Printf("failed to delete PipelineRun %s/%s, error: %v\n", item.GetName(), namespace, delErr)
			} else {
				toDelete--
				fmt.Printf("ok to delete PipelineRun %s/%s\n", item.GetName(), namespace)
			}
		}
	}
	return
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
			return false
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
	return
}

func okToDelete(object map[string]interface{}, maxAge time.Duration) bool {
	if completionTime, err := getCompletionTimeFromObject(object); err == nil {
		return completionTime.Add(maxAge).Before(time.Now())
	}
	return false
}

func (o *gcOption) runE(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("gc failed in %d namespaces: %v", len(errorsNs), errorsNs)
	}
	return nil
}
