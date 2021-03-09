package component

import (
	"context"
	"github.com/linuxsuren/ks/kubectl-plugin/common"
	kstypes "github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
)

type killOption struct {
	client dynamic.Interface

	namespace string
	name      string
}

func newComponentsKillCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := killOption{
		client: client,
	}
	cmd = &cobra.Command{
		Use:               "kill",
		Short:             "Kill the pods of the components",
		Example:           "ks com kill apiserver",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: common.ArrayCompletion("apiserver", "controller", "console", "jenkins", "installer"),
		PreRunE:           opt.preRunE,
		RunE:              opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.namespace, "namespace", "", "", "The namespace of the component")
	flags.StringVarP(&opt.namespace, "ns", "", "", "The namespace of the component")
	flags.StringVarP(&opt.name, "name", "", "", "The name of the component")
	return
}

func (o *killOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	if len(args) > 0 {
		o.name = args[0]
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
		o.name = "ks-apiserver"
		o.namespace = "kubesphere-devops-system"
	case "installer":
		o.name = "ks-installer"
		o.namespace = "kubesphere-system"
	}
	return
}

func (o *killOption) runE(cmd *cobra.Command, args []string) (err error) {
	ctx := context.TODO()

	err = o.client.Resource(kstypes.GetPodSchema()).Namespace(o.namespace).DeleteCollection(ctx, metav1.DeleteOptions{
		TypeMeta: metav1.TypeMeta{
			Kind: "pod",
		},
	}, metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": o.name,
		}).String(),
	})
	return
}
