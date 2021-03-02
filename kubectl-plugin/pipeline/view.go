package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

// NewPipelineViewCmd returns a command to view pipeline
func NewPipelineViewCmd(client dynamic.Interface) (cmd *cobra.Command) {
	ctx := context.TODO()
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
					if rawPip, err = client.Resource(types.GetPipelineSchema()).Namespace(ns).Get(ctx, pip, metav1.GetOptions{}); err == nil {
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
