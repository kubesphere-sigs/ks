package install

import "github.com/spf13/cobra"

// NewInstallCmd returns the command of install kubesphere
func NewInstallCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "install",
		Short: "Install KubeSphere with kind or kk",
	}

	cmd.AddCommand(newInstallWithKindCmd(),
		newInstallWithKKCmd(),
		newInstallerCmd(),
		newInstallK3DCmd())
	return
}
