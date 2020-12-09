package main

import (
	"context"
	"fmt"
	pkg "github.com/linuxsuren/cobra-extension"
	extver "github.com/linuxsuren/cobra-extension/version"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

// NewCmdKS returns the root command of kubeclt-ks
func NewCmdKS(streams genericclioptions.IOStreams) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "ks",
		Short: `kubectl plugin for Kubesphere
Kubesphere is the enterprise-grade container platform tailored for multicloud and multi-cluster management
See also https://github.com/kubesphere/kubesphere`,
	}

	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	var config *rest.Config
	var err error
	var client dynamic.Interface

	if config, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err != nil {
		panic(err)
		return
	}
	if client, err = dynamic.NewForConfig(config); err != nil {
		panic(err)
		return
	}

	cmd.AddCommand(NewUserCmd(client),
		NewPipelineCmd(client),
		extver.NewVersionCmd("linuxsuren", "ks", "kubectl-ks", nil),
		pkg.NewCompletionCmd(cmd))
	return
}

// NewUserCmd returns the command of users
func NewUserCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "user",
		Short: "Reset the password of Kubesphere to the default value which is same with its name",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]

			_, err = client.Resource(GetUserSchema()).Patch(context.TODO(),
				name,
				types.MergePatchType,
				[]byte(fmt.Sprintf(`{"spec":{"password":"%s"},"metadata":{"annotations":null}}`, name)),
				metav1.PatchOptions{})
			return
		},
	}
	return
}
