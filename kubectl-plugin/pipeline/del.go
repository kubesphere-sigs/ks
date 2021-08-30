package pipeline

import (
	"context"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

// newDelPipelineCmd returns a command to delete pipelines
func newDelPipelineCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"del", "remove", "rm"},
		Short:   "Delete a specific Pipeline of KubeSphere DevOps",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var pips []string
			var ns string
			if ns, pips, err = getPipelinesWithConfirm(client, args); err == nil {
				for _, pip := range pips {
					fmt.Println(pip)
					if err = client.Resource(types.GetPipelineSchema()).Namespace(ns).Delete(context.TODO(), pip, metav1.DeleteOptions{}); err != nil {
						break
					}
				}
			}
			return
		},
	}
	return
}
