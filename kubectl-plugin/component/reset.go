package component

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	kstypes "github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"time"
)

// NewComponentResetCmd returns a command to enable (or disable) a component by name
func NewComponentResetCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &ResetOption{
		Option: Option{
			Client: client,
		},
	}
	cmd = &cobra.Command{
		Use:     "reset",
		Short:   "reset the component by name",
		Example: `'kubectl ks com reset -r=false -a' will reset ks-apiserver, ks-controller-manager, ks-console to the latest
kubectl ks com reset -a --nightly latest`,
		RunE:    opt.resetRunE,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opt.Release, "release", "r", true,
		"Indicate if you want to update KubeSphere deploy image to release. Released images come from KubeSphere/xxx. Otherwise images come from kubespheredev/xxx")
	flags.StringVarP(&opt.Tag, "tag", "t", kstypes.KsVersion,
		"The tag of KubeSphere deploys")
	flags.BoolVarP(&opt.ResetAll, "all", "a", false,
		"Indicate if you want to all supported components")
	flags.StringVarP(&opt.Nightly, "nightly", "", "",
		"Indicate if you want to update component to nightly build. It should be date, e.g. 2021-01-01. Or you can just use latest represents the last day")
	flags.StringVarP(&opt.Name, "name", "n", "",
		"The name of target component which you want to reset. This does not work if you provide flag --all")
	return
}

func (o *ResetOption) resetRunE(cmd *cobra.Command, args []string) (err error) {
	if o.Tag == "" {
		// let users choose it if the tag option is empty
		dc := kstypes.DockerClient{
			Image: "kubesphere/ks-apiserver",
		}

		var tags *kstypes.DockerTags
		if tags, err = dc.GetTags(); err != nil {
			err = fmt.Errorf("cannot get the tags, %#v", err)
			return
		}

		prompt := &survey.Select{
			Message: "Please select the tag which you want to check:",
			Options: tags.Tags,
		}
		if err = survey.AskOne(prompt, &o.Tag); err != nil {
			return
		}
	}

	imageOrg := "kubespheredev"
	if o.Release {
		imageOrg = "kubesphere"
	} else {
		o.Tag = "latest"
	}

	// try to parse the nightly date
	if o.Nightly == "latest" {
		imageOrg = "kubespheredev"
		o.Tag = fmt.Sprintf("nightly-%s", time.Now().AddDate(0, 0, -1).Format("20060102"))
	} else if o.Nightly != "" {
		layout := "2006-01-02"
		var targetDate time.Time
		if targetDate, err = time.Parse(layout, o.Nightly); err == nil {
			o.Tag = targetDate.Format("20060102")
		}
	}

	if o.ResetAll {
		o.Name = "apiserver"
		if err = o.updateBy(imageOrg, o.Tag); err != nil {
			return
		}

		o.Name = "controller"
		if err = o.updateBy(imageOrg, o.Tag); err != nil {
			return
		}

		o.Name = "console"
		if err = o.updateBy(imageOrg, o.Tag); err != nil {
			return
		}
	} else {
		err = o.updateBy(imageOrg, o.Tag)
	}
	return
}
