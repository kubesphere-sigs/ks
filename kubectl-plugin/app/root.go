package app

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

func NewAppCmd(client dynamic.Interface, clientset *kubernetes.Clientset) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "app",
		Short: "Manage applications as the GitOps way",
	}

	cmd.AddCommand(newUpdateCmd(client, clientset))
	return
}
