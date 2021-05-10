package install

import (
	"fmt"
	"github.com/linuxsuren/http-downloader/pkg/installer"
	"github.com/linuxsuren/ks/kubectl-plugin/common"
	"github.com/spf13/cobra"
	"strings"
)

const (
	// DefaultKubeSphereVersion is the default version of KubeSphere
	DefaultKubeSphereVersion = "v3.1.0"
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
	flags.StringVarP(&opt.version, "version", "v", DefaultKubeSphereVersion,
		fmt.Sprintf("The version of KubeSphere. Support value could be %s, nightly, nightly-20210309. nightly equals to nightly-latest", DefaultKubeSphereVersion))
	flags.StringArrayVarP(&opt.components, "components", "", []string{},
		"The components which you want to enable after the installation")
	flags.StringVarP(&opt.zone, "zone", "", "cn",
		"Set environment variables, for example export KKZONE=cn")
	return
}

type kkOption struct {
	version           string
	kubernetesVersion string
	components        []string
	zone              string
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
	} else if o.version != DefaultKubeSphereVersion {
		switch o.version {
		case DefaultKubeSphereVersion, "v3.0.0":
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

	is := installer.Installer{
		Provider: "github",
	}
	err = is.CheckDepAndInstall(map[string]string{
		"kk": "kubesphere/kubekey",
	})
	return
}

func (o *kkOption) runE(cmd *cobra.Command, args []string) (err error) {
	report := installReport{}
	report.init()

	commander := Commander{
		Env: []string{fmt.Sprintf("KKZONE=%s", o.zone)},
	}
	if err = commander.execCommand("kk", "create", "cluster", "--with-kubesphere", o.version, "--yes"); err != nil {
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
