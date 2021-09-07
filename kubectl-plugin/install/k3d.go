package install

import (
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/linuxsuren/http-downloader/pkg/installer"
	"github.com/spf13/cobra"
	"runtime"
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
	flags.StringVarP(&opt.image, "image", "", "rancher/k3s:v1.18.20-k3s1",
		"The image of k3s, get more images from https://hub.docker.com/r/rancher/k3s/tags")
	flags.StringVarP(&opt.registry, "registry", "r", "registry",
		"Connect to one or more k3d-managed registries running locally")
	flags.BoolVarP(&opt.withKubeSphere, "with-kubesphere", "", true,
		"Indicate if install KubeSphere as well")
	flags.BoolVarP(&opt.withKubeSphere, "with-ks", "", true,
		"Indicate if install KubeSphere as well")

	// TODO find a better way to reuse the flags from another command
	flags.StringVarP(&opt.version, "version", "", types.KsVersion,
		"The version of KubeSphere which you want to install")
	flags.StringVarP(&opt.nightly, "nightly", "", "",
		"The nightly version you want to install")
	flags.StringArrayVarP(&opt.components, "components", "", []string{},
		"The components that you want to Enabled with KubeSphere")
	flags.BoolVarP(&opt.fetch, "fetch", "", true,
		"Indicate if fetch the latest config of tools")
	return
}

type k3dOption struct {
	installerOption

	image    string
	name     string
	agents   int
	servers  int
	registry string
}

func (o *k3dOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	if o.name == "" && len(args) > 0 {
		o.name = args[0]
	}

	is := installer.Installer{
		Provider: "github",
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Fetch:    o.fetch,
	}
	err = is.CheckDepAndInstall(map[string]string{
		"k3d":     "rancher/k3d",
		"docker":  "docker",
		"kubectl": "kubectl",
	})
	return
}

func (o *k3dOption) runE(cmd *cobra.Command, args []string) (err error) {
	freePort := &common.FreePort{}
	var ports []int
	if ports, err = freePort.FindFreePortsOfKubeSphere(); err != nil {
		return
	}

	// always to create a registry to make sure it's exist
	_ = common.ExecCommand("k3d", "registry", "create", o.registry)

	k3dArgs := []string{"cluster", "create",
		"-p", fmt.Sprintf(`%d:30880@agent[0]`, ports[0]),
		"-p", fmt.Sprintf(`%d:30180@agent[0]`, ports[1]),
		"--agents", fmt.Sprintf("%d", o.agents),
		"--servers", fmt.Sprintf("%d", o.servers),
		"--image", o.image,
		"--registry-use", o.registry}
	if o.name != "" {
		k3dArgs = append(k3dArgs, o.name)
	}
	err = common.ExecCommand("k3d", k3dArgs...)
	return
}

func (o *k3dOption) postRunE(cmd *cobra.Command, args []string) (err error) {
	if !o.withKubeSphere {
		// no need to continue due to no require for KubeSphere
		return
	}

	if err = o.installerOption.preRunE(cmd, args); err == nil {
		err = o.installerOption.runE(cmd, args)
	}
	return
}
