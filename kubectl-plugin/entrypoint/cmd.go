package entrypoint

import (
	"fmt"
	pkg "github.com/linuxsuren/cobra-extension"
	extver "github.com/linuxsuren/cobra-extension/version"
	"github.com/linuxsuren/ks/kubectl-plugin/auth"
	"github.com/linuxsuren/ks/kubectl-plugin/common"
	"github.com/linuxsuren/ks/kubectl-plugin/component"
	"github.com/linuxsuren/ks/kubectl-plugin/config"
	"github.com/linuxsuren/ks/kubectl-plugin/install"
	"github.com/linuxsuren/ks/kubectl-plugin/pipeline"
	"github.com/linuxsuren/ks/kubectl-plugin/registry"
	"github.com/linuxsuren/ks/kubectl-plugin/source2image"
	token2 "github.com/linuxsuren/ks/kubectl-plugin/token"
	"github.com/linuxsuren/ks/kubectl-plugin/tool"
	"github.com/linuxsuren/ks/kubectl-plugin/update"
	"github.com/linuxsuren/ks/kubectl-plugin/user"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// NewCmdKS returns the root command of kubeclt-ks
func NewCmdKS(streams genericclioptions.IOStreams) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "ks",
		Short: `kubectl plugin for Kubesphere
Kubesphere is the enterprise-grade container platform tailored for multicloud and multi-cluster management
See also https://github.com/kubesphere/kubesphere`,
	}

	var err error
	var client dynamic.Interface
	var clientSet *kubernetes.Clientset
	if client, clientSet, err = common.GetClient(); err != nil {
		fmt.Printf("failed to init the k8s client: %v\n", err)
	}

	cmd.AddCommand(user.NewUserCmd(client),
		pipeline.NewPipelineCmd(client),
		update.NewUpdateCmd(client),
		extver.NewVersionCmd("linuxsuren", "ks", "kubectl-ks", nil),
		pkg.NewCompletionCmd(cmd),
		component.NewComponentCmd(client, clientSet),
		token2.NewTokenCmd(client, clientSet),
		registry.NewRegistryCmd(client),
		auth.NewAuthCmd(client),
		tool.NewToolCmd(),
		install.NewInstallCmd(),
		config.NewConfigRootCmd(client),
		source2image.NewS2ICmd(client))
	return
}
