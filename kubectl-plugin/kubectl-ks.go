package main

import (
	"github.com/linuxsuren/ks/kubectl-plugin/entrypoint"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-ks", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := entrypoint.NewCmdKS(genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
