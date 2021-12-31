package option

import (
	"bytes"
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/Masterminds/sprig"
	"github.com/Pallinder/go-randomdata"
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

// PipelineCreateOption is the option for creating a pipeline
type PipelineCreateOption struct {
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

// Wizard is the wizard for creating a pipeline
func (o *PipelineCreateOption) Wizard(_ *cobra.Command, _ []string) (err error) {
	if o.Batch {
		// without Wizard in batch mode
		return
	}

	if o.Workspace == "" {
		var wsNames []string
		if wsNames, err = o.GetWorkspaceTemplateNameList(); err == nil {
			if o.Workspace, err = ChooseObjectFromArray("workspace name", wsNames); err != nil {
				return
			}
		} else {
			return
		}
	}

	if o.Project == "" {
		var projectNames []string
		if projectNames, err = o.getDevOpsProjectNameList(); err == nil {
			if o.Project, err = ChooseObjectFromArray("project name", projectNames); err != nil {
				return
			}
		} else {
			return
		}
	}

	if o.Template == "" {
		if o.Template, err = ChooseOneFromArray(tpl.GetAllTemplates()); err != nil {
			return
		}
	}

	if o.Name == "" {
		defaultVal := fmt.Sprintf("%s-%s", o.Template, strings.ToLower(randomdata.SillyName()))
		if o.Name, err = getInput("Please input the Pipeline name", defaultVal); err != nil {
			return
		}
	}
	return
}

// ChooseObjectFromArray chooses a object from array
func ChooseObjectFromArray(object string, options []string) (result string, err error) {
	prompt := &survey.Select{
		Message: fmt.Sprintf("Please select %s:", object),
		Options: options,
	}
	err = survey.AskOne(prompt, &result)
	return
}

// ChooseOneFromArray choose an item from array
func ChooseOneFromArray(options []string) (result string, err error) {
	result, err = ChooseObjectFromArray("", options)
	return
}

func getInput(title, defaultVal string) (result string, err error) {
	prompt := &survey.Input{
		Message: title,
		Default: defaultVal,
	}
	err = survey.AskOne(prompt, &result)
	return
}

// ParseTemplate parses a template
func (o *PipelineCreateOption) ParseTemplate() (err error) {
	switch o.Template {
	case "":
	case "java":
		o.Jenkinsfile = tpl.GetBuildJava()
	case "go":
		o.Jenkinsfile = tpl.GetBuildGo()
	case "simple":
		o.Jenkinsfile = tpl.GetSimple()
	case "parameter":
		o.Jenkinsfile = tpl.GetParameter()
	case "longRun":
		o.Jenkinsfile = tpl.GetLongRunPipeline()
	case "parallel":
		o.Jenkinsfile = tpl.GetParallel()
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
	return
}

// CreatePipeline creates pipeline
func (o *PipelineCreateOption) CreatePipeline() (err error) {
	ctx := context.TODO()

	var wdID string
	if !o.SkipCheck {
		var ws *unstructured.Unstructured
		if ws, err = o.CheckWorkspace(); err != nil {
			return
		}
		wdID = string(ws.GetUID())
	}

	var project *unstructured.Unstructured
	if project, err = o.CheckDevOpsProject(wdID); err != nil {
		err = fmt.Errorf("cannot find devopsProject %s, error %v", o.Project, err)
		return
	}
	o.Project = project.GetName() // the previous name is a generated name

	var rawPip *unstructured.Unstructured
	if rawPip, err = o.createPipelineObj(); err == nil {
		if rawPip, err = o.Client.Resource(types.GetPipelineSchema()).Namespace(o.Project).Create(ctx, rawPip, metav1.CreateOptions{}); err != nil {
			err = fmt.Errorf("failed to create Pipeline, namespace is %s, error is %v", o.Project, err)
		}
	}
	return
}

// GetDevOpsNamespaceList returns a DevOps namespace list
func (o *PipelineCreateOption) GetDevOpsNamespaceList() (names []string, err error) {
	names, err = o.getUnstructuredNameList(true, []string{}, types.GetDevOpsProjectSchema())
	return
}

func (o *PipelineCreateOption) getDevOpsProjectNameList() (names []string, err error) {
	names, err = o.getUnstructuredNameList(false, []string{}, types.GetDevOpsProjectSchema())
	return
}

func (o *PipelineCreateOption) getWorkspaceNameList() (names []string, err error) {
	names, err = o.getUnstructuredNameList(true, []string{"system-workspace"}, types.GetWorkspaceSchema())
	return
}

// GetWorkspaceTemplateNameList returns a template name list
func (o *PipelineCreateOption) GetWorkspaceTemplateNameList() (names []string, err error) {
	names, err = o.getUnstructuredNameList(true, []string{"system-workspace"}, types.GetWorkspaceTemplate())
	return
}

// GetUnstructuredNameListInNamespace returns a list
func (o *PipelineCreateOption) GetUnstructuredNameListInNamespace(namespace string, originalName bool, excludes []string, schemaType schema.GroupVersionResource) (names []string, err error) {
	var wsList *unstructured.UnstructuredList
	if namespace != "" {
		wsList, err = o.GetUnstructuredListInNamespace(namespace, schemaType)
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

func (o *PipelineCreateOption) getUnstructuredNameList(originalName bool, excludes []string, schemaType schema.GroupVersionResource) (names []string, err error) {
	return o.GetUnstructuredNameListInNamespace("", originalName, excludes, schemaType)
}

// GetUnstructuredListInNamespace returns the list
func (o *PipelineCreateOption) GetUnstructuredListInNamespace(namespace string, schemaType schema.GroupVersionResource) (
	wsList *unstructured.UnstructuredList, err error) {
	ctx := context.TODO()
	wsList, err = o.Client.Resource(schemaType).Namespace(namespace).List(ctx, metav1.ListOptions{})
	return
}

func (o *PipelineCreateOption) getUnstructuredList(schemaType schema.GroupVersionResource) (wsList *unstructured.UnstructuredList, err error) {
	ctx := context.TODO()
	wsList, err = o.Client.Resource(schemaType).List(ctx, metav1.ListOptions{})
	return
}

func (o *PipelineCreateOption) getWorkspaceList() (wsList *unstructured.UnstructuredList, err error) {
	wsList, err = o.getUnstructuredList(types.GetWorkspaceSchema())
	return
}

func (o *PipelineCreateOption) getWorkspaceTemplateList() (wsList *unstructured.UnstructuredList, err error) {
	wsList, err = o.getUnstructuredList(types.GetWorkspaceTemplate())
	return
}

// CheckWorkspace makes sure the target workspace exist
func (o *PipelineCreateOption) CheckWorkspace() (ws *unstructured.Unstructured, err error) {
	ctx := context.TODO()
	if ws, err = o.Client.Resource(types.GetWorkspaceSchema()).Get(ctx, o.Workspace, metav1.GetOptions{}); err == nil {
		return
	}

	// TODO check workspaceTemplate when ks in a multi-cluster environment
	if ws, err = o.Client.Resource(types.GetWorkspaceTemplate()).Get(ctx, o.Workspace, metav1.GetOptions{}); err != nil {
		// create workspacetemplate
		var wsTemplate *unstructured.Unstructured
		if wsTemplate, err = types.GetObjectFromYaml(fmt.Sprintf(`apiVersion: tenant.kubesphere.io/v1alpha2
kind: WorkspaceTemplate
metadata:
  name: %s`, o.Workspace)); err != nil {
			err = fmt.Errorf("failed to unmarshal yaml to DevOpsProject object, %v", err)
			return
		}

		ws, err = o.Client.Resource(types.GetWorkspaceTemplate()).Create(ctx, wsTemplate, metav1.CreateOptions{})
	}
	return
}

func (o *PipelineCreateOption) getDevOpsProjectGenerateNameList() (names []string, err error) {
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

func (o *PipelineCreateOption) getDevOpsProjectList() (wsList *unstructured.UnstructuredList, err error) {
	ctx := context.TODO()
	selector := labels.Set{"kubesphere.io/workspace": o.Workspace}
	wsList, err = o.Client.Resource(types.GetDevOpsProjectSchema()).List(ctx, metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(selector).String(),
	})
	return
}

// CheckDevOpsProject makes sure the project exist
func (o *PipelineCreateOption) CheckDevOpsProject(wsID string) (project *unstructured.Unstructured, err error) {
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

func (o *PipelineCreateOption) createPipelineObj() (rawPip *unstructured.Unstructured, err error) {
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
