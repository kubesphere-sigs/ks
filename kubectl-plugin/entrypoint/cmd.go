package entrypoint

import (
	"context"
	"fmt"
	pkg "github.com/linuxsuren/cobra-extension"
	extver "github.com/linuxsuren/cobra-extension/version"
	"github.com/linuxsuren/ks/kubectl-plugin/auth"
	"github.com/linuxsuren/ks/kubectl-plugin/component"
	"github.com/linuxsuren/ks/kubectl-plugin/install"
	"github.com/linuxsuren/ks/kubectl-plugin/pipeline"
	"github.com/linuxsuren/ks/kubectl-plugin/registry"
	token2 "github.com/linuxsuren/ks/kubectl-plugin/token"
	"github.com/linuxsuren/ks/kubectl-plugin/tool"
	kstype "github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/linuxsuren/ks/kubectl-plugin/update"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
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
	var clientSet *kubernetes.Clientset

	if config, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err != nil {
		fmt.Println(err)
	} else {
		if client, err = dynamic.NewForConfig(config); err != nil {
			fmt.Println(err)
		}

		if clientSet, err = kubernetes.NewForConfig(config); err != nil {
			fmt.Println(err)
		}
	}

	cmd.AddCommand(NewUserCmd(client),
		pipeline.NewPipelineCmd(client),
		update.NewUpdateCmd(client),
		extver.NewVersionCmd("linuxsuren", "ks", "kubectl-ks", nil),
		pkg.NewCompletionCmd(cmd),
		component.NewComponentCmd(client, clientSet),
		token2.NewTokenCmd(client, clientSet),
		registry.NewRegistryCmd(client),
		auth.NewAuthCmd(client),
		tool.NewToolCmd(),
		install.NewInstallCmd())
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

			_, err = client.Resource(kstype.GetUserSchema()).Patch(context.TODO(),
				name,
				types.MergePatchType,
				[]byte(fmt.Sprintf(`{"spec":{"password":"%s"},"metadata":{"annotations":null}}`, name)),
				metav1.PatchOptions{})
			return
		},
	}
	return
}
