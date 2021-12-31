package uninstall

import (
	"context"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	component2 "github.com/kubesphere-sigs/ks/kubectl-plugin/common/component"
	kstypes "github.com/kubesphere-sigs/ks/kubectl-plugin/types"
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
		Short:   "Uninstall Component Of KubeSphere",
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
	GetName() string
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

		switch component {
		case "alerting", "auditing", "devops", "events", "logging", "metrics_server", "networkpolicy", "openpitrix", "servicemesh":
			if component == "openpitrix" {
				patch = fmt.Sprintf(`[{"op": "replace", "path": "/spec/openpitrix.store.enabled", "value": %s}]`, strconv.FormatBool(false))
			}
		default:
			err = fmt.Errorf("not support [%s] yet", component)
			return
		}
		comp := o.getComponent(component)

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

func (o *uninstallOption) getComponent(name string) Component {
	allComponents := make(map[string]Component)

	alerting := &component2.Alerting{}
	allComponents[alerting.GetName()] = alerting

	auditing := &component2.Auditing{}
	allComponents[auditing.GetName()] = auditing

	devops := &component2.DevOps{Client: o.Client, Clientset: o.Clientset}
	allComponents[devops.GetName()] = devops

	events := &component2.Events{}
	allComponents[events.GetName()] = events

	logging := &component2.Logging{}
	allComponents[logging.GetName()] = logging

	metricServer := &component2.MetricsServer{}
	allComponents[metricServer.GetName()] = metricServer

	networkPolicy := &component2.NetworkPolicy{}
	allComponents[networkPolicy.GetName()] = networkPolicy

	openPitrix := &component2.OpenPitrix{}
	allComponents[openPitrix.GetName()] = openPitrix

	serviceMesh := &component2.ServiceMesh{}
	allComponents[serviceMesh.GetName()] = serviceMesh

	return allComponents[name]
}
