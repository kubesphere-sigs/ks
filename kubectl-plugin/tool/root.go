package tool

import (
	"fmt"
	verup "github.com/linuxsuren/cobra-extension/version"
	"github.com/spf13/cobra"
)

// NewToolCmd creates the tool command
func NewToolCmd() (cmd *cobra.Command) {
	opt := &option{}
	cmd = &cobra.Command{
		Use:     "tool",
		Short:   "Install tools from KubeSphere family, such as: kk, ke",
		Example: "kubectl ks tool kk",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}
	return
}

func (o *option) preRunE(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 {
		err = fmt.Errorf("tool name is necessary")
		return
	}

	if len(args) == 1 {
		o.Org = "kubesphere"
		o.Name = args[0]

		switch o.Name {
		case "kk":
			o.Name = "kk"
			o.Repo = "kubekey"
		case "ke":
			o.Name = "ke"
			o.Repo = "kubeye"
		}
	} else if len(args) >= 2 {
		o.Org = args[0]
		o.Name = args[1]
	}

	if o.Repo == "" {
		o.Repo = o.Name
	}
	return
}

func (o *option) runE(cmd *cobra.Command, args []string) (err error) {
	opt := &verup.SelfUpgradeOption{
		Org:  o.Org,
		Name: o.Name,
		Repo: o.Repo,
	}

	err = opt.Download(cmd, "", "", fmt.Sprintf("/usr/local/bin/%s", o.Name))
	return
}

type option struct {
	Org  string
	Name string
	Repo string
}
