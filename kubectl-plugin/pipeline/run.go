package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"text/template"
)

func newPipelineRunCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &pipelineRunOpt{
		client:               client,
		pipelineCreateOption: pipelineCreateOption{Client: client},
	}
	cmd = &cobra.Command{
		Use:     "run",
		Short:   "Start a Pipeline",
		Long:    "Start a Pipeline. Only v1alpha4 is supported",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.pipeline, "pipeline", "p", "",
		"The Pipeline name that you want to run")
	flags.StringVarP(&opt.namespace, "namespace", "n", "",
		"The namespace of target Pipeline")
	flags.BoolVarP(&opt.batch, "batch", "b", false, "Run pipeline as batch mode")
	return
}

type pipelineRunOpt struct {
	pipeline  string
	namespace string
	batch     bool

	// inner fields
	client dynamic.Interface
	pipelineCreateOption
}

func (o *pipelineRunOpt) preRunE(cmd *cobra.Command, args []string) (err error) {
	if o.pipeline == "" && len(args) > 0 {
		o.pipeline = args[0]
	}

	if err = o.wizard(cmd, args); err != nil {
		return
	}
	return
}

func (o *pipelineRunOpt) runE(_ *cobra.Command, _ []string) (err error) {
	var tpl *template.Template
	if tpl, err = template.New("pipelineRunTpl").Parse(pipelineRunTpl); err != nil {
		err = fmt.Errorf("failed to parse template:'%s', error: %v", pipelineRunTpl, err)
		return
	}

	var buf bytes.Buffer
	if err = tpl.Execute(&buf, map[string]string{
		"name":      o.pipeline,
		"namespace": o.namespace,
	}); err != nil {
		err = fmt.Errorf("failed render pipeline template, error: %v", err)
		return
	}

	var pipelineRunObj *unstructured.Unstructured
	if pipelineRunObj, err = types.GetObjectFromYaml(buf.String()); err != nil {
		err = fmt.Errorf("failed to unmarshal yaml to DevOpsProject object, %v", err)
		return
	}

	if _, err = o.client.Resource(types.GetPipelineRunSchema()).Namespace(o.namespace).Create(context.TODO(),
		pipelineRunObj, metav1.CreateOptions{}); err != nil {
		err = fmt.Errorf("failed create PipelineRun, error: %v", err)
	}
	return
}

func (o *pipelineRunOpt) wizard(_ *cobra.Command, _ []string) (err error) {
	if o.batch {
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

	if o.namespace == "" {
		var projectNames []string
		if projectNames, err = o.getDevOpsNamespaceList(); err == nil {
			if o.namespace, err = chooseObjectFromArray("project name", projectNames); err != nil {
				return
			}
		} else {
			return
		}
	}

	if o.pipeline == "" {
		var pipelineNames []string
		if pipelineNames, err = o.getPipelineNameList(); err == nil && len(pipelineNames) > 0 {
			if o.pipeline, err = chooseObjectFromArray("pipeline name", pipelineNames); err != nil {
				return
			}
		} else if len(pipelineNames) == 0 {
			err = fmt.Errorf("no Pipelines found in namespace '%s', please create it first", o.namespace)
			return
		}
	}
	return
}

func (o *pipelineRunOpt) getPipelineNameList() (names []string, err error) {
	names, err = o.getUnstructuredNameListInNamespace(o.namespace, true, []string{}, types.GetPipelineSchema())
	return
}

var pipelineRunTpl = `
apiVersion: devops.kubesphere.io/v1alpha4
kind: PipelineRun
metadata:
  generateName: {{.name}}
  namespace: {{.namespace}}
spec:
  pipelineRef:
    name: {{.name}}
`
