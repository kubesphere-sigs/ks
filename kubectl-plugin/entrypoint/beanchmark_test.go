package entrypoint_test

import (
	"bytes"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/entrypoint"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
	"testing"
)

func BenchmarkHelps(b *testing.B) {
	rootCmd := entrypoint.NewCmdKS(genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})
	runCmdHelps(b, rootCmd)
}

func runCmdHelps(b *testing.B, cmd *cobra.Command) {
	b.Run(entrypoint.GetCmdPath(cmd), func(b *testing.B) {
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetArgs([]string{"--help"})
		_, _ = cmd.ExecuteC()
	})

	// run sub-commands
	for i := range cmd.Commands() {
		runCmdHelps(b, cmd.Commands()[i])
	}
}
