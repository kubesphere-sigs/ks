package entrypoint_test

import (
	"bytes"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/entrypoint"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"testing"
)

func TestNewCmdKSInit(t *testing.T) {
	cmd := entrypoint.NewCmdKS(genericclioptions.IOStreams{})
	assert.NotNil(t, t, cmd, "failed to init root command")

	testCmdHelp(t, cmd)
}

func testCmdHelp(t *testing.T, cmd *cobra.Command) {
	t.Run(entrypoint.GetCmdPath(cmd), func(t *testing.T) {
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetArgs([]string{"--help"})
		_, _ = cmd.ExecuteC()
	})

	// run sub-commands
	for i := range cmd.Commands() {
		testCmdHelp(t, cmd.Commands()[i])
	}
}
