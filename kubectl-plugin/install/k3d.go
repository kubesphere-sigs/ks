package install

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/linuxsuren/http-downloader/pkg/installer"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"text/template"
)

func newInstallK3DCmd() (cmd *cobra.Command) {
	opt := &k3dOption{}
	cmd = &cobra.Command{
		Use:   "k3d",
		Short: "Install KubeSphere with k3d",
		Long: `Install KubeSphere with k3d
You can get more details from https://github.com/rancher/k3d/


[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
    endpoint = ["http://k3d-registry:5000"]

cat /var/lib/rancher/k3s/agent/etc/containerd/config.toml.tmpl


quay.io
`,
		PreRunE:  opt.preRunE,
		RunE:     opt.runE,
		PostRunE: opt.postRunE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.name, "name", "n", "k3s-default",
		"The name of k3d cluster")
	flags.IntVarP(&opt.agents, "agents", "", 1,
		"Specify how many agents you want to create")
	flags.IntVarP(&opt.servers, "servers", "", 1,
		"Specify how many servers you want to create")
	flags.StringVarP(&opt.image, "image", "", "rancher/k3s:v1.19.14-k3s1",
		"The image of k3s, get more images from https://hub.docker.com/r/rancher/k3s/tags")
	flags.StringVarP(&opt.registry, "registry", "r", "registry",
		"Connect to one or more k3d-managed registries running locally")
	flags.BoolVarP(&opt.withKubeSphere, "with-kubesphere", "", true,
		"Indicate if install KubeSphere as well")
	flags.BoolVarP(&opt.withKubeSphere, "with-ks", "", true,
		"Indicate if install KubeSphere as well")
	flags.BoolVarP(&opt.reInstall, "reinstall", "", false,
		"Indicate if re-install the k3d cluster with given name")
	flags.IntVarP(&opt.extraFreePorts, "extra-free-ports", "", 0,
		"Open more extra free ports for the k3d, this flag is the count of the extra free ports instead of the value")

	// TODO find a better way to reuse the flags from another command
	flags.StringVarP(&opt.version, "version", "", types.KsVersion,
		"The version of KubeSphere which you want to install")
	flags.StringVarP(&opt.nightly, "nightly", "", "",
		"The nightly version you want to install")
	flags.StringArrayVarP(&opt.components, "components", "", []string{},
		"The components that you want to Enabled with KubeSphere")
	flags.BoolVarP(&opt.fetch, "fetch", "", true,
		"Indicate if fetch the latest config of tools")

	// completion for flags
	_ = cmd.RegisterFlagCompletionFunc("image", common.ArrayCompletion("rancher/k3s:v1.19.14-k3s1",
		"rancher/k3s:v1.20.10-k3s1",
		"rancher/k3s:v1.21.4-k3s1"))
	_ = cmd.RegisterFlagCompletionFunc("components", common.PluginAbleComponentsCompletion())
	_ = cmd.RegisterFlagCompletionFunc("nightly", common.ArrayCompletion("latest"))
	return
}

type k3dOption struct {
	installerOption

	image          string
	name           string
	agents         int
	servers        int
	registry       string
	reInstall      bool
	extraFreePorts int
}

func (o *k3dOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	if len(args) > 0 {
		o.name = args[0]
	}

	// make the name of nightly k3d be clear
	if o.name == "k3s-default" && o.nightly != "" {
		_, o.name = common.GetNightlyTag(o.nightly)
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
	freePort := common.NewFreePort(o.extraFreePorts)
	var ports []int
	if ports, err = freePort.FindFreePortsOfKubeSphere(); err != nil {
		return
	}

	// always to create registries to make sure they are exist
	registryMap := createRegistries(o.registry)

	if o.reInstall {
		_ = common.ExecCommand("k3d", "cluster", "delete", o.name)
	}

	agentPort, err := getAgentPort()
	if err != nil {
		return err
	}

	var freePortList []string
	for i := range ports {
		port := ports[i]
		freePortList = append(freePortList, "-p", fmt.Sprintf(`%d:%d@%s`, port, port, agentPort))
	}

	cfgFile := getRegistryConfigFile(registryMap)

	k3dArgs := []string{"cluster", "create",
		"--agents", fmt.Sprintf("%d", o.agents),
		"--servers", fmt.Sprintf("%d", o.servers),
		"--image", o.image,
		"--registry-config", cfgFile,
		o.name}
	for reg := range registryMap {
		k3dArgs = append(k3dArgs, "--registry-use", reg)
	}

	k3dArgs = append(k3dArgs, freePortList...)
	err = common.ExecCommand("k3d", k3dArgs...)
	return
}

func getRegistryConfigFile(regMap map[string]string) (targetFile string) {
	data := []byte(getRegistryConfig(regMap))
	targetFile = os.ExpandEnv("$HOME/.registry.yaml")
	_ = ioutil.WriteFile(targetFile, data, 0644)
	return
}

func getRegistryConfig(regMap map[string]string) string {
	tpl, err := template.New("mirror").Parse(`mirrors:
{{- range $key, $value := .}}
  {{$value}}:
    endpoint:
    - http://k3d-{{$key}}:5000
  k3d-{{$key}}:5000:
    endpoint:
    - http://k3d-{{$key}}:5000
{{- end}}
configs: {}
auths: {}`)
	if err != nil {
		return ""
	}
	buf := bytes.NewBuffer([]byte{})
	if err = tpl.Execute(buf, regMap); err != nil {
		return ""
	}
	return buf.String()
}

func createRegistries(prefix string) (regMap map[string]string) {
	regMap = make(map[string]string, 3)
	regMap[createRegistry(prefix, "")] = "docker.io"
	regMap[createRegistry(prefix, "ghcr")] = "ghcr.io"
	regMap[createRegistry(prefix, "quay")] = "quay.io"
	return
}

func createRegistry(prefix, name string) (alias string) {
	alias = prefix
	tagPrefix := ""
	if name != "" {
		alias = alias + "-" + name
		tagPrefix = name + "-"
	}
	_ = common.ExecCommand("k3d", "registry", "create", alias, "--image",
		fmt.Sprintf("ghcr.io/kubesphere-sigs/registry:%s2", tagPrefix))
	return
}

//getAgentPort get the agent port string via local command `k3d version`
func getAgentPort() (string, error) {
	out, err := common.ExecCommandGetOutput("k3d", "version")
	if err != nil {
		return "", err
	}

	isNewVersion, err := isGreaterThanV5(out)
	if err != nil {
		return "", err
	}

	// if k3d version is greater than v5+
	if isNewVersion {
		return "agent:0", nil
	}
	// adaptation before k3d v4
	return "agent[0]", nil
}

//isGreaterThanV5 check if k3d version is greater than v5
func isGreaterThanV5(version string) (bool, error) {
	c, _ := semver.NewConstraint(">= 5.0.0")
	reg := regexp.MustCompile(`(\w+\.){2}\w+`)
	if reg != nil {
		raw := reg.FindAllStringSubmatch(version, 1)
		v, err := semver.NewVersion(raw[0][0])
		if err != nil {
			return false, fmt.Errorf("Error parsing version: %s", err.Error())
		}
		return c.Check(v), nil
	}
	return false, nil
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
