package main

import (
	"fmt"
	ext "github.com/linuxsuren/cobra-extension"
	extver "github.com/linuxsuren/cobra-extension/version"
	aliasCmd "github.com/linuxsuren/go-cli-alias/pkg/cmd"
	"github.com/linuxsuren/ks/kubectl-plugin/entrypoint"
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
	cmd.AddCommand(extver.NewVersionCmd("linuxsuren", AliasCLI, AliasCLI, nil))

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
		cmd.AddCommand(kubectlPluginCmds...)
	}

	aliasCmd.Execute(cmd, targetCommand, getDefault(), nil)
}
