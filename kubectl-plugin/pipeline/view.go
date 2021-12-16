package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

// newPipelineViewCmd returns a command to view pipeline
func newPipelineViewCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "view",
		Short: "Output the YAML format of a Pipeline",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var pips []string
			var ns string
			if ns, pips, err = getPipelinesWithConfirm(client, args); err == nil {
				for _, pip := range pips {
					var rawPip *unstructured.Unstructured
					var data []byte
					buf := bytes.NewBuffer(data)
					if rawPip, err = getPipeline(pip, ns, client); err == nil {
						enc := json.NewEncoder(buf)
						enc.SetIndent("", "    ")
						if err = enc.Encode(rawPip); err != nil {
							return
						}
					} else {
						err = fmt.Errorf("cannot get pipeline, error: %#v", err)
						return
					}

					var yamlData []byte
					if yamlData, err = yaml.JSONToYAML(buf.Bytes()); err != nil {
						return
					}

					cmd.Println(string(yamlData))
				}
			}
			return
		},
	}
	return
}

func getPipeline(name, namespace string, client dynamic.Interface) (rawPip *unstructured.Unstructured, err error) {
	rawPip, err = client.Resource(types.GetPipelineSchema()).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	return
}
