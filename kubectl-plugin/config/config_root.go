package config

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
)

// NewConfigRootCmd returns the config command
func NewConfigRootCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "option",
		Short:   "Config KubeSphere as you need",
		Aliases: []string{"opt"},
	}

	cmd.AddCommand(newClusterCmd(client), newMigrateCmd(client))
	return
}
