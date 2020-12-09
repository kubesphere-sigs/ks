package main

import (
	ext "github.com/linuxsuren/cobra-extension"
	extver "github.com/linuxsuren/cobra-extension/version"
	aliasCmd "github.com/linuxsuren/go-cli-alias/pkg/cmd"
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

	aliasCmd.Execute(cmd, TargetCLI, getDefault(), nil)
}
