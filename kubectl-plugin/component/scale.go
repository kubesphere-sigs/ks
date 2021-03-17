package component

import (
	"fmt"
	"github.com/linuxsuren/ks/kubectl-plugin/common"
	"github.com/spf13/cobra"
)

func newScaleCmd() (cmd *cobra.Command) {
	opt := &scaleOption{}

	cmd = &cobra.Command{
		Use:     "scale",
		Short:   "Set a new size for a Deployment, ReplicaSet, Replication Controller, or StatefulSet of the KubeSphere components",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.IntVarP(&opt.replicas, "replicas", "r", -1, "The new desired number of replicas")
	flags.StringVarP(&opt.name, "name", "n", "", "The name of KubeSphere component")
	return
}

type scaleOption struct {
	name      string
	namespace string
	replicas  int
}

func (o *scaleOption) preRunE(_ *cobra.Command, args []string) (err error) {
	if len(args) > 0 {
		o.name = args[0]
	}

	if o.name == "" {
		err = fmt.Errorf("provide the name of component")
		return
	}

	if o.replicas < 0 {
		err = fmt.Errorf("the value of replicas must be >= 0")
		return
	}

	switch o.name {
	case "apiserver":
		o.name = "ks-apiserver"
		o.namespace = "kubesphere-system"
	case "controller":
		o.name = "ks-controller-manager"
		o.namespace = "kubesphere-system"
	case "console":
		o.name = "ks-console"
		o.namespace = "kubesphere-system"
	case "jenkins":
		o.name = "ks-jenkins"
		o.namespace = "kubesphere-devops-system"
	case "installer":
		o.name = "ks-installer"
		o.namespace = "kubesphere-system"
	}
	return
}

func (o *scaleOption) runE(_ *cobra.Command, _ []string) (err error) {
	err = common.ExecCommand("kubectl", "scale", "-n", o.namespace, "--replicas",
		fmt.Sprintf("%d", o.replicas), fmt.Sprintf("deploy/%s", o.name))
	return
}
