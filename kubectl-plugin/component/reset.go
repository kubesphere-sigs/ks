package component

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	kstypes "github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
)

// NewComponentResetCmd returns a command to enable (or disable) a component by name
func NewComponentResetCmd() (cmd *cobra.Command) {
	opt := &ResetOption{}
	cmd = &cobra.Command{
		Use:   "reset",
		Short: "Reset the component by name",
		Example: `'ks com reset -r=false --nightly latest console' will reset ks-console to the latest release
'ks com reset -r=false -a' will reset ks-apiserver, ks-controller-manager, ks-console to the latest
'ks com reset -a --nightly latest' will reset the images to the latest nightly which comes from yesterday
'ks com reset -a --nightly latest-1' will reset the images to the nightly which comes from the day before yesterday`,
		PreRunE: opt.preRunE,
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

func (o *ResetOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	ctx := cmd.Root().Context()
	o.Client = common.GetDynamicClient(ctx)
	o.Clientset = common.GetClientset(ctx)

	if o.Name == "" && len(args) > 0 {
		o.Name = args[0]
	}
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
	if o.Release && o.Nightly == "" {
		imageOrg = "kubesphere"
	} else if o.Tag == "" {
		// try to parse the nightly date
		_, o.Tag = common.GetNightlyTag(o.Nightly)
	}

	if o.ResetAll {
		o.Name = "apiserver"
		if err = o.updateBy(imageOrg); err != nil {
			return
		}

		o.Name = "controller"
		if err = o.updateBy(imageOrg); err != nil {
			return
		}

		o.Name = "console"
		if err = o.updateBy(imageOrg); err != nil {
			return
		}

		o.Name = "installer"
		if err = o.updateBy(imageOrg); err != nil {
			return
		}
	} else {
		if o.Name == "" {
			err = fmt.Errorf("please provide a component name")
		} else {
			err = o.updateBy(imageOrg)
		}
	}
	return
}
