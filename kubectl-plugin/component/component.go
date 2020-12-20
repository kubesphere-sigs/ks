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
		Short:   "Manage the components of KubeSphere",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cmd.Println("show all components status here. it's best to offer some options to filter these components")
			return
		},
	}

	cmd.AddCommand(NewComponentEnableCmd(client),
		NewComponentEditCmd(client),
		NewComponentResetCmd(client),
		NewComponentWatchCmd(client),
		NewComponentLogCmd(client))
	return
}

type Option struct {
	Name string
}

// NewComponentEnableCmd returns a command to enable (or disable) a component by name
func NewComponentEnableCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "enable",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return
}

// NewComponentWatchCmd returns a command to enable (or disable) a component by name
func NewComponentWatchCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &Option{}
	cmd = &cobra.Command{
		Use: "watch",
		RunE: opt.watchRunE,
	}
	return
}

func (o *Option) watchRunE(cmd *cobra.Command, args []string) (err error) {
	return nil
}

// NewComponentResetCmd returns a command to enable (or disable) a component by name
func NewComponentResetCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "reset",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return
}

// NewComponentLogCmd returns a command to enable (or disable) a component by name
func NewComponentLogCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "log",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return
}

// NewComponentEditCmd returns a command to enable (or disable) a component by name
func NewComponentEditCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "edit",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return
}
