package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"text/template"
)

func newPipelineRunCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &pipelineRunOpt{
		client: client,
	}
	cmd = &cobra.Command{
		Use:   "run",
		Short: "Start a Pipeline",
		Long:  "Start a Pipeline. Only v1alpha4 is supported",
		RunE:  opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.pipeline, "pipeline", "p", "",
		"The Pipeline name that you want to run")
	flags.StringVarP(&opt.namespace, "namespace", "n", "",
		"The namespace of target Pipeline")
	return
}

type pipelineRunOpt struct {
	pipeline  string
	namespace string

	// inner fields
	client dynamic.Interface
}

func (o *pipelineRunOpt) runE(cmd *cobra.Command, args []string) (err error) {
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
