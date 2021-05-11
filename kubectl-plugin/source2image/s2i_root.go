package source2image

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
)

// NewS2ICmd creates a command to manage image builder in KubeSphere
func NewS2ICmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "s2i",
		Short: "Manage image builder in KubeSphere",
	}

	cmd.AddCommand(createS2i(client))
	return
}
