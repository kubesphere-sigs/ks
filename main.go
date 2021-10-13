package main

import (
	"context"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/entrypoint"
	ext "github.com/linuxsuren/cobra-extension/pkg"
	extver "github.com/linuxsuren/cobra-extension/version"
	aliasCmd "github.com/linuxsuren/go-cli-alias/pkg/cmd"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
)

const (
	// TargetCLI represents target CLI which is kubectl
	TargetCLI = "kubectl"
	// AliasCLI represents the alias CLI which is ks
	AliasCLI = "ks"
	// KubernetesType is the env name for Kubernetes type
	KubernetesType = "kubernetes_type"
)

func main() {
	kType := os.Getenv(KubernetesType)
	var cmd *cobra.Command
	var targetCommand string

	switch kType {
	case "k3s":
		targetCommand = fmt.Sprintf("k3s %s", TargetCLI)
	default:
		targetCommand = TargetCLI
	}

	cmd = aliasCmd.CreateDefaultCmd(TargetCLI, AliasCLI)
	cmd.AddCommand(extver.NewVersionCmd("kubesphere-sigs", AliasCLI, AliasCLI, nil))

	aliasCmd.AddAliasCmd(cmd, getDefault())
	cmd.AddCommand(ext.NewCompletionCmd(cmd))

	// need to figure out how to connect with k3s before enable below features
	if kType != "k3s" {
		// add all the sub-commands from kubectl-ks
		kubectlPluginCmdRoot := entrypoint.NewCmdKS(genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		})
		kubectlPluginCmds := kubectlPluginCmdRoot.Commands()
		cmd.PersistentFlags().AddFlagSet(kubectlPluginCmdRoot.PersistentFlags())
		cmd.PersistentPreRunE = kubectlPluginCmdRoot.PersistentPreRunE
		cmd.AddCommand(kubectlPluginCmds...)
	}

	aliasCmd.ExecuteContext(cmd, context.WithValue(context.TODO(), common.ClientFactory{}, &common.ClientFactory{}),
		targetCommand, getDefault(), nil)
}
