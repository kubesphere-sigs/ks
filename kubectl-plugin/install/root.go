package install

import "github.com/spf13/cobra"

// NewInstallCmd returns the command of install kubesphere
func NewInstallCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "install",
		Short: "install KubeSphere",
	}

	cmd.AddCommand(newInstallWithKindCmd())
	return
}
