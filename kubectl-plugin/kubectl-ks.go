package main

import (
	"context"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/entrypoint"
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
	if err := root.ExecuteContext(context.WithValue(context.TODO(), common.ClientFactory{}, &common.ClientFactory{})); err != nil {
		os.Exit(1)
	}
}
