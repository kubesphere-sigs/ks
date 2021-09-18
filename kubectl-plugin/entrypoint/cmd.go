package entrypoint

import (
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/auth"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/component"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/config"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/install"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/registry"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/source2image"
	token2 "github.com/kubesphere-sigs/ks/kubectl-plugin/token"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/tool"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/update"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/user"
	pkg "github.com/linuxsuren/cobra-extension"
	extver "github.com/linuxsuren/cobra-extension/version"
	hdcmd "github.com/linuxsuren/http-downloader/pkg/cmd"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type rootOpt struct {
	context string
}

func (o *rootOpt) persistentPreRunE(cmd *cobra.Command, args []string) (err error) {
	factory := cmd.Context().Value(common.ClientFactory{})
	factory.(*common.ClientFactory).SetContext(o.context)
	return
}

// NewCmdKS returns the root command of kubeclt-ks
func NewCmdKS(streams genericclioptions.IOStreams) (cmd *cobra.Command) {
	opt := &rootOpt{}
	cmd = &cobra.Command{
		Use: "ks",
		Short: `kubectl plugin for KubeSphere
KubeSphere is the enterprise-grade container platform tailored for multicloud and multi-cluster management
See also https://github.com/kubesphere/kubesphere`,
		PersistentPreRunE: opt.persistentPreRunE,
	}
	cmd.PersistentFlags().StringVarP(&opt.context, "context", "c", "", "")

	var err error
	var client dynamic.Interface
	var clientSet *kubernetes.Clientset
	if client, clientSet, err = common.GetClient(); err != nil {
		fmt.Printf("failed to init the k8s client: %v\n", err)
	}

	cmd.AddCommand(user.NewUserCmd(client),
		pipeline.NewPipelineCmd(client),
		update.NewUpdateCmd(client),
		extver.NewVersionCmd("kubesphere-sigs", "ks", "kubectl-ks", nil),
		pkg.NewCompletionCmd(cmd),
		component.NewComponentCmd(client, clientSet),
		token2.NewTokenCmd(client, clientSet),
		registry.NewRegistryCmd(client),
		auth.NewAuthCmd(client),
		tool.NewToolCmd(),
		install.NewInstallCmd(),
		config.NewConfigRootCmd(client),
		source2image.NewS2ICmd(client),
		hdcmd.NewInitCommand(map[string]string{
			"kubectl": "kubectl",
		}, map[string]string{
			"vim": "vim",
		}))
	return
}
