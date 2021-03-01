package main

import (
	ext "github.com/linuxsuren/cobra-extension"
	extver "github.com/linuxsuren/cobra-extension/version"
	aliasCmd "github.com/linuxsuren/go-cli-alias/pkg/cmd"
	"github.com/linuxsuren/ks/kubectl-plugin/entrypoint"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
)

const (
	// TargetCLI represents target CLI which is kubectl
	TargetCLI = "kubectl"
	// AliasCLI represents the alias CLI which is ks
	AliasCLI = "ks"
)

func main() {
	cmd := aliasCmd.CreateDefaultCmd(TargetCLI, AliasCLI)

	cmd.AddCommand(extver.NewVersionCmd("linuxsuren", AliasCLI, AliasCLI, nil))

	aliasCmd.AddAliasCmd(cmd, getDefault())
	cmd.AddCommand(ext.NewCompletionCmd(cmd))

	// add all the sub-commands from kubectl-ks
	kubectlPluginCmdRoot := entrypoint.NewCmdKS(genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})
	kubectlPluginCmds := kubectlPluginCmdRoot.Commands()
	cmd.AddCommand(kubectlPluginCmds...)

	aliasCmd.Execute(cmd, TargetCLI, getDefault(), nil)
}
