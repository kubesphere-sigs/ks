package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"os"
	"sigs.k8s.io/yaml"
)

// NewPipelineEditCmd returns a command to edit the pipeline
func NewPipelineEditCmd(client dynamic.Interface) (cmd *cobra.Command) {
	ctx := context.TODO()
	cmd = &cobra.Command{
		Use:     "edit",
		Aliases: []string{"e"},
		Short:   "Edit the target pipeline",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var pips []string
			var ns string
			if ns, pips, err = getPipelinesWithConfirm(client, args); err == nil {
				for _, pip := range pips {
					var rawPip *unstructured.Unstructured
					var data []byte
					buf := bytes.NewBuffer(data)
					cmd.Printf("get pipeline %s/%s\n", ns, pip)
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

					var fileName = "*.yaml"
					var content = string(yamlData)

					prompt := &survey.Editor{
						Message:       fmt.Sprintf("Edit pipeline %s/%s", ns, pip),
						FileName:      fileName,
						Default:       string(yamlData),
						HideDefault:   true,
						AppendDefault: true,
					}

					err = survey.AskOne(prompt, &content, survey.WithStdio(os.Stdin, os.Stdout, os.Stderr))

					if err = yaml.Unmarshal([]byte(content), rawPip); err == nil {
						_, err = client.Resource(types.GetPipelineSchema()).Namespace(ns).Update(context.TODO(), rawPip, metav1.UpdateOptions{})
					}
				}
			}
			return
		},
	}
	return
}

func getPipelinesWithConfirm(client dynamic.Interface, args []string) (ns string, pips []string, err error) {
	var allPips []string
	if ns, allPips, err = getPipelines(client, args); err != nil {
		return
	}

	prompt := &survey.MultiSelect{
		Message: "Please select the pipelines that you want to check:",
		Options: allPips,
	}
	err = survey.AskOne(prompt, &pips)
	return
}

func getPipelines(client dynamic.Interface, args []string) (ns string, pips []string, err error) {
	if len(args) >= 2 {
		ns, pips = args[0], args[1:]
		return
	} else if len(args) == 1 {
		ns = args[0]
	} else {
		if ns, err = getNamespace(client, args); err != nil {
			return
		}
	}

	var list *unstructured.UnstructuredList
	if list, err = client.Resource(types.GetPipelineSchema()).Namespace(ns).List(context.TODO(), metav1.ListOptions{}); err == nil {
		for _, item := range list.Items {
			pips = append(pips, item.GetName())
		}
	}
	return
}
