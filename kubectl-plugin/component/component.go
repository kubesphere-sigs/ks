package component

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
)

// NewComponentCmd returns a command to manage components of Kubesphere
func NewComponentCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "component",
		Aliases: []string{"com"},
		Short: "Manage the components of Kubesphere",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cmd.Println("show all components status here. it's best to offer some options to filter these components")
			return
		},
	}

	cmd.AddCommand(NewComponentEnableCmd(client))
	return
}

// NewComponentEnableCmd returns a command to enable (or disable) a component by name
func NewComponentEnableCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{}
	return
}
