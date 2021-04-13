package install

import (
	"fmt"
	"github.com/linuxsuren/ks/kubectl-plugin/common"
	"github.com/spf13/cobra"
)

func newInstallK3DCmd() (cmd *cobra.Command) {
	opt := &k3dOption{}
	cmd = &cobra.Command{
		Use:   "k3d",
		Short: "Install KubeSphere with k3d",
		Long: `Install KubeSphere with k3d
You can get more details from https://github.com/rancher/k3d/`,
		PreRunE:  opt.preRunE,
		RunE:     opt.runE,
		PostRunE: opt.postRunE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.name, "name", "n", "",
		"The name of k3d cluster")
	flags.IntVarP(&opt.agents, "agents", "", 1,
		"Specify how many agents you want to create")
	flags.IntVarP(&opt.servers, "servers", "", 1,
		"Specify how many servers you want to create")

	// TODO find a better way to reuse the flags from another command
	flags.StringVarP(&opt.version, "version", "", "v3.0.0",
		"The version of KubeSphere which you want to install")
	flags.StringVarP(&opt.nightly, "nightly", "", "",
		"The nightly version you want to install")
	flags.StringArrayVarP(&opt.components, "components", "", []string{},
		"The components that you want to Enabled with KubeSphere")
	return
}

type k3dOption struct {
	installerOption

	name    string
	agents  int
	servers int
}

func (o *k3dOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	if o.name == "" && len(args) > 0 {
		o.name = args[0]
	}
	return
}

func (o *k3dOption) runE(cmd *cobra.Command, args []string) (err error) {
	freePort := &common.FreePort{}
	var ports []int
	if ports, err = freePort.FindFreePortsOfKubeSphere(); err != nil {
		return
	}

	k3dArgs := []string{"cluster", "create",
		"-p", fmt.Sprintf(`%d:30880@agent[0]`, ports[0]),
		"-p", fmt.Sprintf(`%d:30180@agent[0]`, ports[1]),
		"--agents", fmt.Sprintf("%d", o.agents),
		"--servers", fmt.Sprintf("%d", o.servers)}
	if o.name != "" {
		k3dArgs = append(k3dArgs, o.name)
	}
	err = common.ExecCommand("k3d", k3dArgs...)
	return
}

func (o *k3dOption) postRunE(cmd *cobra.Command, args []string) (err error) {
	if err = o.installerOption.preRunE(cmd, args); err == nil {
		err = o.installerOption.runE(cmd, args)
	}
	return
}
