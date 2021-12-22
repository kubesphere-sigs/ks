package uninstall

import (
	"context"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	kstypes "github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	component2 "github.com/kubesphere-sigs/ks/kubectl-plugin/uninstall/component"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"strconv"
)

// NewUninstallCmd returns the command of uninstall kubesphere
func NewUninstallCmd() (cmd *cobra.Command) {
	opt := &uninstallOption{}
	cmd = &cobra.Command{
		Use:     "uninstall",
		Short:   "Uninstall KubeSphere",
		Example: `ks uninstall --components devops`,
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.StringSliceVarP(&opt.components, "components", "", []string{},
		"Which components will uninstall")

	_ = cmd.RegisterFlagCompletionFunc("components", common.PluginAbleComponentsCompletion())
	return
}

type uninstallOption struct {
	components []string

	Client    dynamic.Interface
	Clientset *kubernetes.Clientset
}

// Component return the interface of Component
type Component interface {
	Uninstall() error
}

func (o *uninstallOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	ctx := cmd.Root().Context()
	o.Client = common.GetDynamicClient(ctx)
	o.Clientset = common.GetClientset(ctx)
	return
}

func (o *uninstallOption) runE(cmd *cobra.Command, args []string) (err error) {
	ctx := context.TODO()

	ns, name := "kubesphere-system", "ks-installer"
	for _, component := range o.components {
		patch := fmt.Sprintf(`[{"op": "replace", "path": "/spec/%s/enabled", "value": %s}]`, component, strconv.FormatBool(false))

		var comp Component
		switch component {
		case "alerting":
			comp = &component2.Alerting{}
			break
		case "auditing":
			comp = &component2.Auditing{}
			break
		case "devops":
			comp = &component2.DevOps{}
			break
		case "events":
			comp = &component2.Events{}
			break
		case "logging":
			comp = &component2.Logging{}
			break
		case "metrics_server":
			comp = &component2.MetricsServer{}
			break
		case "networkpolicy":
			comp = &component2.NetworkPolicy{}
			break
		case "openpitrix":
			patch = fmt.Sprintf(`[{"op": "replace", "path": "openpitrix.store.enabled", "value": %s}]`, strconv.FormatBool(false))
			comp = &component2.OpenPitrix{}
			break
		case "servicemesh":
			patch = fmt.Sprintf(`[{"op": "replace", "path": "servicemesh.enabled", "value": %s}]`, strconv.FormatBool(false))
			comp = &component2.ServiceMesh{}
			break
		default:
			err = fmt.Errorf("not support [%s] yet", component)
			return
		}

		_, err = o.Client.Resource(kstypes.GetClusterConfiguration()).Namespace(ns).Patch(ctx,
			name, types.JSONPatchType,
			[]byte(patch),
			metav1.PatchOptions{})
		if err = comp.Uninstall(); err != nil {
			return err
		}
	}

	return
}
