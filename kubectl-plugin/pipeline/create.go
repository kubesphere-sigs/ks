package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/Masterminds/sprig"
	"github.com/Pallinder/go-randomdata"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/tpl"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	"html/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"strings"
)

type pipelineCreateOption struct {
	Workspace   string
	Project     string
	Name        string
	Jenkinsfile string
	Template    string
	Type        string
	SCMType     string
	Batch       bool
	SkipCheck   bool

	// Inner fields
	Client       dynamic.Interface
	WorkspaceUID string
}

func newPipelineCreateCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &pipelineCreateOption{
		Client: client,
	}

	cmd = &cobra.Command{
		Use:   "create",
		Short: "Create a Pipeline in the KubeSphere cluster",
		Long: `Create a Pipeline in the KubeSphere cluster
You can create a Pipeline with a java, go template. Before you do that, please make sure the workspace exists.
KubeSphere supports multiple types Pipeline. Currently, this CLI only support the simple one with Jenkinsfile inside.'`,
		Example: "ks pip create --ws simple --template java --name java --project test",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.Workspace, "workspace", "", "",
		"The workspace name of KubeSphere cluster")
	flags.StringVarP(&opt.Workspace, "ws", "", "",
		"The workspace name of KubeSphere cluster. This is an alias for --workspace")
	flags.StringVarP(&opt.Project, "project", "", "",
		"The DevOps project name of KubeSphere cluster")
	flags.StringVarP(&opt.Name, "name", "", "",
		"The name of the Pipeline")
	flags.StringVarP(&opt.Jenkinsfile, "jenkinsfile", "", "",
		"The Jenkinsfile of the Pipeline")
	flags.StringVarP(&opt.Template, "template", "", "",
		"Template of Jenkinsfile include: java, go. This option will override the option --jenkinsfile")
	flags.StringVarP(&opt.Type, "type", "", "pipeline",
		"The type of pipeline, could be pipeline, multi_branch_pipeline")
	flags.StringVarP(&opt.SCMType, "scm-type", "", "",
		"The SCM type of pipeline, could be gitlab, github")
	flags.BoolVarP(&opt.Batch, "batch", "b", false, "Create pipeline as batch mode")
	flags.BoolVarP(&opt.SkipCheck, "skip-check", "", false, "Skip the resources check")

	_ = cmd.RegisterFlagCompletionFunc("template", common.ArrayCompletion("java", "go", "simple", "longRun",
		"multi-branch-gitlab", "multi-branch-github", "multi-branch-git"))
	_ = cmd.RegisterFlagCompletionFunc("type", common.ArrayCompletion("pipeline", "multi-branch-pipeline"))
	_ = cmd.RegisterFlagCompletionFunc("scm-type", common.ArrayCompletion("gitlab", "github", "git"))

	// TODO needs to find a better way to add the completion support
	// it takes long time (around 1min) to initialize the whole command if the k8s config is not reachable
	//if client != nil {
	//	// these features rely on the k8s client, ignore it if the client is nil
	//	if wsList, err := opt.getWorkspaceNameList(); err == nil {
	//		_ = cmd.RegisterFlagCompletionFunc("ws", common.ArrayCompletion(wsList...))
	//	}
	//	if projectList, err := opt.getDevOpsProjectGenerateNameList(); err == nil {
	//		_ = cmd.RegisterFlagCompletionFunc("project", common.ArrayCompletion(projectList...))
	//	}
	//}
	return
}

func (o *pipelineCreateOption) wizard(_ *cobra.Command, _ []string) (err error) {
	if o.Batch {
		// without wizard in batch mode
		return
	}

	if o.Workspace == "" {
		var wsNames []string
		if wsNames, err = o.getWorkspaceTemplateNameList(); err == nil {
			if o.Workspace, err = chooseObjectFromArray("workspace name", wsNames); err != nil {
				return
			}
		} else {
			return
		}
	}

	if o.Project == "" {
		var projectNames []string
		if projectNames, err = o.getDevOpsProjectNameList(); err == nil {
			if o.Project, err = chooseObjectFromArray("project name", projectNames); err != nil {
				return
			}
		} else {
			return
		}
	}

	if o.Template == "" {
		if o.Template, err = chooseOneFromArray([]string{"java", "go", "simple", "longRun",
			"multi-branch-gitlab", "multi-branch-github", "multi-branch-git"}); err != nil {
			return
		}
	}

	if o.Name == "" {
		if o.Name, err = getInput("Please input the Pipeline name"); err != nil {
			return
		}
	}
	return
}

func chooseObjectFromArray(object string, options []string) (result string, err error) {
	prompt := &survey.Select{
		Message: fmt.Sprintf("Please select %s:", object),
		Options: options,
	}
	err = survey.AskOne(prompt, &result)
	return
}

func chooseOneFromArray(options []string) (result string, err error) {
	result, err = chooseObjectFromArray("", options)
	return
}

func getInput(title string) (result string, err error) {
	prompt := &survey.Input{
		Message: title,
		Default: strings.ToLower(randomdata.SillyName()),
	}
	err = survey.AskOne(prompt, &result)
	return
}

func (o *pipelineCreateOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	if o.Name == "" && len(args) > 0 {
		o.Name = args[0]
	}

	if err = o.wizard(cmd, args); err != nil {
		return
	}

	switch o.Template {
	case "":
	case "java":
		o.Jenkinsfile = tpl.GetBuildJava()
	case "go":
		o.Jenkinsfile = tpl.GetBuildGo()
	case "simple":
		o.Jenkinsfile = tpl.GetSimple()
	case "longRun":
		o.Jenkinsfile = tpl.GetLongRunPipeline()
	case "multi-branch-git":
		o.Type = "multi-branch-pipeline"
		o.SCMType = "git"
	case "multi-branch-gitlab":
		o.Type = "multi-branch-pipeline"
		o.SCMType = "gitlab"
	case "multi-branch-github":
		o.Type = "multi-branch-pipeline"
		o.SCMType = "github"
	default:
		err = fmt.Errorf("%s is not support", o.Template)
	}
	o.Jenkinsfile = strings.TrimSpace(o.Jenkinsfile)

	if o.Name == "" && len(args) > 0 {
		o.Name = args[0]
	}

	if o.Name == "" {
		err = fmt.Errorf("please provide the name of Pipeline")
	}
	return
}

func (o *pipelineCreateOption) runE(cmd *cobra.Command, args []string) (err error) {
	ctx := context.TODO()

	var wdID string
	if !o.SkipCheck {
		var ws *unstructured.Unstructured
		if ws, err = o.checkWorkspace(); err != nil {
			return
		}
		wdID = string(ws.GetUID())
	}

	var project *unstructured.Unstructured
	if project, err = o.checkDevOpsProject(wdID); err != nil {
		return
	}
	o.Project = project.GetName() // the previous name is the generate name

	var rawPip *unstructured.Unstructured
	if rawPip, err = o.createPipelineObj(); err == nil {
		if rawPip, err = o.Client.Resource(types.GetPipelineSchema()).Namespace(o.Project).Create(ctx, rawPip, metav1.CreateOptions{}); err != nil {
			err = fmt.Errorf("failed to create Pipeline, %v", err)
		}
	}
	return
}

func (o *pipelineCreateOption) getDevOpsNamespaceList() (names []string, err error) {
	names, err = o.getUnstructuredNameList(true, []string{}, types.GetDevOpsProjectSchema())
	return
}

func (o *pipelineCreateOption) getDevOpsProjectNameList() (names []string, err error) {
	names, err = o.getUnstructuredNameList(false, []string{}, types.GetDevOpsProjectSchema())
	return
}

func (o *pipelineCreateOption) getWorkspaceNameList() (names []string, err error) {
	names, err = o.getUnstructuredNameList(true, []string{"system-workspace"}, types.GetWorkspaceSchema())
	return
}

func (o *pipelineCreateOption) getWorkspaceTemplateNameList() (names []string, err error) {
	names, err = o.getUnstructuredNameList(true, []string{"system-workspace"}, types.GetWorkspaceTemplate())
	return
}

func (o *pipelineCreateOption) getUnstructuredNameListInNamespace(namespace string, originalName bool, excludes []string, schemaType schema.GroupVersionResource) (names []string, err error) {
	var wsList *unstructured.UnstructuredList
	if namespace != "" {
		wsList, err = o.getUnstructuredListInNamespace(namespace, schemaType)
	} else {
		wsList, err = o.getUnstructuredList(schemaType)
	}

	if err == nil {
		names = make([]string, 0)
		for i := range wsList.Items {
			var name string
			if originalName {
				name = wsList.Items[i].GetName()
			} else {
				name = wsList.Items[i].GetGenerateName()
			}

			exclude := false
			for j := range excludes {
				if name == excludes[j] {
					exclude = true
					break
				}
			}

			if !exclude {
				names = append(names, name)
			}
		}
	}
	return
}

func (o *pipelineCreateOption) getUnstructuredNameList(originalName bool, excludes []string, schemaType schema.GroupVersionResource) (names []string, err error) {
	return o.getUnstructuredNameListInNamespace("", originalName, excludes, schemaType)
}

func (o *pipelineCreateOption) getUnstructuredListInNamespace(namespace string, schemaType schema.GroupVersionResource) (
	wsList *unstructured.UnstructuredList, err error) {
	ctx := context.TODO()
	wsList, err = o.Client.Resource(schemaType).Namespace(namespace).List(ctx, metav1.ListOptions{})
	return
}

func (o *pipelineCreateOption) getUnstructuredList(schemaType schema.GroupVersionResource) (wsList *unstructured.UnstructuredList, err error) {
	ctx := context.TODO()
	wsList, err = o.Client.Resource(schemaType).List(ctx, metav1.ListOptions{})
	return
}

func (o *pipelineCreateOption) getWorkspaceList() (wsList *unstructured.UnstructuredList, err error) {
	wsList, err = o.getUnstructuredList(types.GetWorkspaceSchema())
	return
}

func (o *pipelineCreateOption) getWorkspaceTemplateList() (wsList *unstructured.UnstructuredList, err error) {
	wsList, err = o.getUnstructuredList(types.GetWorkspaceTemplate())
	return
}

func (o *pipelineCreateOption) checkWorkspace() (ws *unstructured.Unstructured, err error) {
	ctx := context.TODO()
	if ws, err = o.Client.Resource(types.GetWorkspaceSchema()).Get(ctx, o.Workspace, metav1.GetOptions{}); err == nil {
		return
	}

	// TODO check workspaceTemplate when ks in a multi-cluster environment
	ws, err = o.Client.Resource(types.GetWorkspaceTemplate()).Get(ctx, o.Workspace, metav1.GetOptions{})
	return
}

func (o *pipelineCreateOption) getDevOpsProjectGenerateNameList() (names []string, err error) {
	var list *unstructured.UnstructuredList
	if list, err = o.getDevOpsProjectList(); err != nil {
		return
	}

	names = make([]string, len(list.Items))
	for i := range list.Items {
		names[i] = list.Items[i].GetGenerateName()
	}
	return
}

func (o *pipelineCreateOption) getDevOpsProjectList() (wsList *unstructured.UnstructuredList, err error) {
	ctx := context.TODO()
	selector := labels.Set{"kubesphere.io/workspace": o.Workspace}
	wsList, err = o.Client.Resource(types.GetDevOpsProjectSchema()).List(ctx, metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(selector).String(),
	})
	return
}

func (o *pipelineCreateOption) checkDevOpsProject(wsID string) (project *unstructured.Unstructured, err error) {
	ctx := context.TODO()
	var list *unstructured.UnstructuredList
	if list, err = o.getDevOpsProjectList(); err != nil {
		return
	}

	found := false
	for i := range list.Items {
		if list.Items[i].GetGenerateName() == o.Project {
			found = true
			project = &list.Items[i]
			break
		}
	}

	if !found {
		var tpl *template.Template
		o.WorkspaceUID = wsID
		if tpl, err = template.New("project").Parse(devopsProjectTemplate); err != nil {
			err = fmt.Errorf("failed to parse devops project template, error is: %v", err)
			return
		}

		var buf bytes.Buffer
		if err = tpl.Execute(&buf, o); err != nil {
			return
		}

		var projectObj *unstructured.Unstructured
		if projectObj, err = types.GetObjectFromYaml(buf.String()); err != nil {
			err = fmt.Errorf("failed to unmarshal yaml to DevOpsProject object, %v", err)
			return
		}

		if project, err = o.Client.Resource(types.GetDevOpsProjectSchema()).Create(ctx, projectObj, metav1.CreateOptions{}); err != nil {
			err = fmt.Errorf("failed to create devops project with YAML: '%s'. Error is: %v", buf.String(), err)
		}
	}
	return
}

func (o *pipelineCreateOption) createPipelineObj() (rawPip *unstructured.Unstructured, err error) {
	var tpl *template.Template
	funcMap := sprig.FuncMap()
	//funcMap["raw"] = html.UnescapeString
	funcMap["raw"] = func(text string) template.HTML {
		/* #nosec */
		return template.HTML(text)
	}
	if tpl, err = template.New("pipeline").Funcs(funcMap).Parse(pipelineTemplate); err != nil {
		err = fmt.Errorf("failed to parse Pipeline template, %v", err)
		return
	}

	var buf bytes.Buffer
	if err = tpl.Execute(&buf, o); err != nil {
		err = fmt.Errorf("failed to render Pipeline template, %v", err)
		return
	}

	if rawPip, err = types.GetObjectFromYaml(buf.String()); err != nil {
		err = fmt.Errorf("failed to unmarshal yaml to Pipeline object, %v", err)
	}
	return
}

var devopsProjectTemplate = `
apiVersion: devops.kubesphere.io/v1alpha3
kind: DevOpsProject
metadata:
  annotations:
    kubesphere.io/creator: admin
  finalizers:
  - devopsproject.finalizers.kubesphere.io
  generateName: {{.Project}}
  labels:
    kubesphere.io/workspace: {{.Workspace}}
  {{if ne .WorkspaceUID ""}}
  ownerReferences:
  - apiVersion: tenant.kubesphere.io/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Workspace
    name: {{.Workspace}}
    uid: {{.WorkspaceUID}}
  {{end}}
`

var pipelineTemplate = `
apiVersion: devops.kubesphere.io/v1alpha3
kind: Pipeline
metadata:
  annotations:
    kubesphere.io/creator: admin
  finalizers:
  - pipeline.finalizers.kubesphere.io
  name: "{{.Name}}"
  namespace: {{.Project}}
spec:
  {{if eq .Type "pipeline"}}
  pipeline:
    disable_concurrent: true
    discarder:
      days_to_keep: "7"
      num_to_keep: "10"
    jenkinsfile: |
{{.Jenkinsfile | indent 6 | raw}}
    name: "{{.Name}}"
  {{else if eq .Type "multi-branch-pipeline" -}}
  multi_branch_pipeline:
    discarder:
      days_to_keep: "-1"
      num_to_keep: "-1"
    {{if eq .SCMType "gitlab"}}
    gitlab_source:
      discover_branches: 1
      discover_pr_from_forks:
        strategy: 2
        trust: 2
      discover_pr_from_origin: 2
      discover_tags: true
      owner: devops-ws
      repo: devops-ws/learn-pipeline-java
      server_name: https://gitlab.com
    {{else if eq .SCMType "github" -}}
    github_source:
      discover_branches: 1
      discover_pr_from_forks:
        strategy: 2
        trust: 2
      discover_pr_from_origin: 2
      discover_tags: true
      owner: devops-ws
      repo: learn-pipeline-java
    {{else if eq .SCMType "git" -}}
    git_source:
      discover_branches: true
      url: https://gitee.com/devops-ws/learn-pipeline-java
    {{end -}}
    name: "{{.Name}}"
    script_path: Jenkinsfile
    source_type: {{.SCMType}}
  {{end -}}
  type: {{.Type}}
status: {}
`
