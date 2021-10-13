package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/option"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"text/template"
)

func newPipelineRunCmd() (cmd *cobra.Command) {
	opt := &pipelineRunOpt{}
	cmd = &cobra.Command{
		Use:     "run",
		Short:   "Start a Pipeline",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.pipeline, "pipeline", "p", "",
		"The Pipeline name that you want to run")
	flags.StringVarP(&opt.namespace, "namespace", "n", "",
		"The namespace of target Pipeline")
	flags.StringVarP(&opt.project, "project", "", "",
		"The project of target Pipeline")
	flags.BoolVarP(&opt.batch, "batch", "b", false, "Run pipeline as batch mode")
	flags.StringToStringVarP(&opt.parameters, "parameters", "P", map[string]string{}, "The parameters that you want to pass, example of single parameter: name=value")
	return
}

type pipelineRunOpt struct {
	pipeline   string
	namespace  string
	project    string
	batch      bool
	parameters map[string]string

	// inner fields
	client dynamic.Interface
	option.PipelineCreateOption
}

func (o *pipelineRunOpt) triggerPipeline(ns, pipeline string, parameters map[string]string) (err error) {
	pipelineRunYaml, err := parsePipelineRunTpl(map[string]interface{}{
		"name":       pipeline,
		"namespace":  ns,
		"parameters": parameters,
	})
	if err != nil {
		return err
	}

	var pipelineRunObj *unstructured.Unstructured
	if pipelineRunObj, err = types.GetObjectFromYaml(pipelineRunYaml); err != nil {
		err = fmt.Errorf("failed to unmarshal yaml to Pipelinerun object, %v", err)
		return
	}

	if _, err = o.client.Resource(types.GetPipelineRunSchema()).Namespace(ns).Create(context.TODO(),
		pipelineRunObj, metav1.CreateOptions{}); err != nil {
		err = fmt.Errorf("failed create PipelineRun, error: %v", err)
	}
	return
}

func (o *pipelineRunOpt) preRunE(cmd *cobra.Command, args []string) (err error) {
	o.client = common.GetDynamicClient(cmd.Root().Context())
	o.PipelineCreateOption.Client = o.client

	if o.pipeline == "" && len(args) > 0 {
		o.pipeline = args[0]
	}

	if o.project != "" {
		var devopsProject *unstructured.Unstructured
		if devopsProject, err = o.getDevOpsProjectByGenerateName(o.project); err == nil && devopsProject != nil {
			o.namespace = devopsProject.GetName()
		} else {
			err = fmt.Errorf("unable to find namespace by devops project: %s, error: %v", o.project, err)
			return
		}
	}

	if err = o.wizard(cmd, args); err != nil {
		return
	}
	return
}

func (o *pipelineRunOpt) runE(_ *cobra.Command, _ []string) (err error) {
	err = o.triggerPipeline(o.namespace, o.pipeline, o.parameters)
	return
}

func (o *pipelineRunOpt) wizard(_ *cobra.Command, _ []string) (err error) {
	if o.batch {
		// without wizard in batch mode
		return
	}

	if o.Workspace == "" {
		var wsNames []string
		if wsNames, err = o.GetWorkspaceTemplateNameList(); err == nil {
			if o.Workspace, err = option.ChooseObjectFromArray("workspace name", wsNames); err != nil {
				return
			}
		} else {
			return
		}
	}

	if o.namespace == "" {
		var projectNames []string
		if projectNames, err = o.GetDevOpsNamespaceList(); err == nil {
			if o.namespace, err = option.ChooseObjectFromArray("project name", projectNames); err != nil {
				return
			}
		} else {
			return
		}
	}

	if o.pipeline == "" {
		var pipelineNames []string
		if pipelineNames, err = o.getPipelineNameList(); err == nil && len(pipelineNames) > 0 {
			if o.pipeline, err = option.ChooseObjectFromArray("pipeline name", pipelineNames); err != nil {
				return
			}
		} else if len(pipelineNames) == 0 {
			err = fmt.Errorf("no Pipelines found in namespace '%s', please create it first", o.namespace)
			return
		}
	}
	return
}

func (o *pipelineRunOpt) getDevOpsProjectByGenerateName(name string) (result *unstructured.Unstructured, err error) {
	ctx := context.TODO()
	var projectList *unstructured.UnstructuredList
	if projectList, err = o.Client.Resource(types.GetDevOpsProjectSchema()).List(ctx, metav1.ListOptions{}); err == nil {
		for i := range projectList.Items {
			item := projectList.Items[i]

			if item.GetGenerateName() == name {
				result = &item
				return
			}
		}
	}
	return
}

func (o *pipelineRunOpt) getPipelineNameList() (names []string, err error) {
	names, err = o.GetUnstructuredNameListInNamespace(o.namespace, true, []string{}, types.GetPipelineSchema())
	return
}

func parsePipelineRunTpl(data map[string]interface{}) (pipelineRunYaml string, err error) {
	var tpl *template.Template
	if tpl, err = template.New("pipelineRunTpl").Parse(pipelineRunTpl); err != nil {
		err = fmt.Errorf("failed to parse template:'%s', error: %v", pipelineRunTpl, err)
		return
	}

	if err != nil {
		return
	}

	var buf bytes.Buffer
	if err = tpl.Execute(&buf, data); err != nil {
		err = fmt.Errorf("failed to render pipeline template, error: %v", err)
		return
	}
	return buf.String(), nil
}

var pipelineRunTpl = `
apiVersion: devops.kubesphere.io/v1alpha3
kind: PipelineRun
metadata:
  generateName: {{.name}}
  namespace: {{.namespace}}
spec:
  pipelineRef:
    name: {{.name}}
  {{- if .parameters }}
  parameters:
	{{- range $name, $value := .parameters }}
    - name: {{ $name | printf "%q" }}
      value: {{ $value | printf "%q" }}
	{{- end }}
  {{- end }}
`
