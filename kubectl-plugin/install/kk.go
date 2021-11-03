package install

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/install/containerd"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/linuxsuren/http-downloader/pkg/exec"
	"github.com/linuxsuren/http-downloader/pkg/installer"
	"github.com/linuxsuren/http-downloader/pkg/net"
	"github.com/spf13/cobra"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func newInstallWithKKCmd() (cmd *cobra.Command) {
	opt := &kkOption{}
	cmd = &cobra.Command{
		Use:     "kk",
		Aliases: []string{"kubekey"},
		Short:   "Install KubeSphere with kubekey (aka kk)",
		Long: `Install KubeSphere with kubekey (aka kk)
Get more details about kubekey from https://github.com/kubesphere/kubekey`,
		Example: `ks install kk --components devops
ks install kk --version nightly --components devops`,
		ValidArgsFunction: common.PluginAbleComponentsCompletion(),
		PreRunE:           opt.preRunE,
		RunE:              opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.version, "version", "v", types.KsVersion,
		fmt.Sprintf("The version of KubeSphere. Support value could be %s, nightly, nightly-20210309. nightly equals to nightly-latest", types.KsVersion))
	flags.StringVarP(&opt.version, "kubernetesVersion", "", types.K8sVersion,
		"The version of Kubernetes")
	flags.StringArrayVarP(&opt.components, "components", "", []string{},
		"The components which you want to enable after the installation")
	flags.StringVarP(&opt.zone, "zone", "", "cn",
		"Set environment variables, for example export KKZONE=cn")
	flags.StringVarP(&opt.container, "container", "", "docker",
		"Indicate the container runtime type. Supported: docker, containerd")
	flags.BoolVarP(&opt.fetch, "fetch", "", true,
		"Indicate if fetch the latest config of tools")

	_ = cmd.RegisterFlagCompletionFunc("components", common.PluginAbleComponentsCompletion())
	_ = cmd.RegisterFlagCompletionFunc("container", common.ArrayCompletion("docker", "containerd"))
	return
}

type kkOption struct {
	version           string
	kubernetesVersion string
	components        []string
	zone              string
	container         string
	fetch             bool
}

func (o *kkOption) versionCheck() (err error) {
	if strings.HasPrefix(o.version, "nightly") {
		ver := strings.ReplaceAll(o.version, "nightly-", "")
		ver = strings.ReplaceAll(ver, "nightly", "")
		if ver == "" {
			ver = "latest"
		}

		if _, ver = common.GetNightlyTag(ver); ver == "" {
			err = fmt.Errorf("not support version: %s", o.version)
		} else {
			o.version = ver
		}
	} else if !isNotReleaseVersion(o.version) && o.version != types.KsVersion {
		switch o.version {
		case types.KsVersion, "v3.0.0":
		default:
			err = fmt.Errorf("not support version: %s", o.version)
		}
	}
	return
}

func (o *kkOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	if err = o.versionCheck(); err != nil {
		return
	}

	is := &installer.Installer{
		Provider: "github",
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Fetch:    o.fetch,
	}
	dep := map[string]string{
		"kk":        "kubesphere/kubekey",
		"socat":     "socat",
		"conntrack": "conntrack",
	}
	switch o.container {
	case "docker":
		dep["docker"] = "docker"
	case "containerd":
		dep["containerd"] = "containerd/containerd"
		dep["crictl"] = "kubernetes-sigs/cri-tools"
		dep["runc"] = "opencontainers/runc"
	}

	if err = is.CheckDepAndInstall(dep); err == nil && o.container == "containerd" {
		err = setDefaultConfigFiles()
	}

	// TODO find a better way to restart service
	if err == nil {
		err = enableAndRestartService(o.container)
	}
	return
}

func enableAndRestartService(service string) (err error) {
	if err = exec.RunCommand("systemctl", "enable", service); err != nil {
		err = fmt.Errorf("failed to enable service: %s, error: %v", service, err)
	} else {
		if err = exec.RunCommand("systemctl", "restart", service); err != nil {
			err = fmt.Errorf("failed to restart service: %s, error: %v", service, err)
		}
	}
	return
}

func setDefaultConfigFiles() (err error) {
	if err = setDefaultIfNotExist([]byte(containerd.GetConfigToml()), "/etc/containerd/config.toml"); err == nil {
		err = setDefaultIfNotExist([]byte(containerd.GetCrictl()), "/etc/crictl.yaml")
	}

	if err != nil {
		return
	}

	err = setDefaultIfNotExist([]byte(containerd.GetContainerdService()), "/etc/systemd/system/containerd.service")
	return
}

func setDefaultIfNotExist(data []byte, path string) (err error) {
	parentDir := filepath.Dir(path)
	if err = os.MkdirAll(parentDir, 0644); err != nil {
		err = fmt.Errorf("failed to create directory: %s, error: %v", parentDir, err)
		return
	}

	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = ioutil.WriteFile(path, data, 0644)
	} else {
		prompt := &survey.Select{
			Message: fmt.Sprintf("If you want to overwrite the existing file '%s'.", path),
			Options: []string{"yes", "no"},
		}
		var choose string
		if err = survey.AskOne(prompt, &choose); err != nil {
			return
		}

		if choose == "yes" {
			err = ioutil.WriteFile(path, data, 0644)
		}
	}
	return
}

func (o *kkOption) runE(cmd *cobra.Command, args []string) (err error) {
	report := installReport{}
	report.init()

	var configFile string
	if configFile, err = getTemporaryConfigFile(o.container); err != nil {
		err = fmt.Errorf("failed to get a temporary kk config file, error: %v", err)
		return
	}

	defer func(file string) {
		_ = os.RemoveAll(file)
	}(configFile)
	commander := Commander{
		Env: []string{fmt.Sprintf("KKZONE=%s", o.zone)},
	}
	if err = commander.execCommand("kk", "create", "cluster", "--filename", configFile,
		"--with-kubesphere", o.version, "--with-kubernetes", o.kubernetesVersion, "--yes"); err != nil {
		return
	}

	for _, component := range o.components {
		if err = commander.execCommand("ks", "com", "enable", component); err != nil {
			return
		}
	}
	report.end()
	return
}

func getTemporaryConfigFile(container string) (filePath string, err error) {
	var (
		kkFile  *os.File
		tpl     *template.Template
		address string
	)

	if kkFile, err = os.CreateTemp(os.TempDir(), "kk-config"); err != nil {
		err = fmt.Errorf("failed to create temporary file for kk config, error: %v", err)
		return
	}

	filePath = kkFile.Name()
	if address, err = net.GetExternalIP(); err != nil {
		err = fmt.Errorf("failed to get external IP, error: %v", err)
		return
	}

	if tpl, err = template.New("config").Parse(containerd.GetKKConfig()); err == nil {
		if err = tpl.Execute(kkFile, map[string]string{
			"container": container,
			"address":   address,
		}); err != nil {
			err = fmt.Errorf("failed to render kk config file template, error: %v", err)
		}
	} else {
		err = fmt.Errorf("failed to create kk config file template, error: %v", err)
	}
	return
}
