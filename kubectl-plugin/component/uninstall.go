package component

import (
	"context"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/component/component"
	kstypes "github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
)

// NewComponentUninstallCmd returns the command of uninstall component of kubesphere
func NewComponentUninstallCmd() (cmd *cobra.Command) {
	opt := &Option{}
	cmd = &cobra.Command{
		Use:     "uninstall",
		Short:   "Uninstall Component Of KubeSphere",
		Example: `You can uninstall a single component with name via: ks com uninstall devops
Or it's possible to uninstall all components via: ks com uninstall all`,
		PreRunE: opt.uninstallPreRunE,
		ValidArgsFunction: common.PluginAbleComponentsCompletion(),
		RunE:    opt.uninstallRunE,
	}

	_ = cmd.RegisterFlagCompletionFunc("components", common.PluginAbleComponentsCompletion())
	return
}

// Component return the interface of Component
type Component interface {
	Uninstall() error
	GetName() string
}

func (o *Option) uninstallPreRunE(cmd *cobra.Command, args []string) (err error) {
	ctx := cmd.Root().Context()
	o.Client = common.GetDynamicClient(ctx)
	o.Clientset = common.GetClientset(ctx)

	return o.componentNameCheck(cmd, args)
}

func (o *Option) uninstallRunE(cmd *cobra.Command, args []string) (err error) {
	ctx := context.TODO()
	ns, name := "kubesphere-system", "ks-installer"
	patch := fmt.Sprintf(`[{"op": "replace", "path": "/spec/%s/enabled", "value": %s}]`, o.Name, strconv.FormatBool(false))

	switch o.Name {
	case "alerting", "auditing", "devops", "events", "logging", "metrics_server", "networkpolicy", "openpitrix", "servicemesh":
		if o.Name == "openpitrix" {
			patch = fmt.Sprintf(`[{"op": "replace", "path": "/spec/openpitrix.store.enabled", "value": %s}]`, strconv.FormatBool(false))
		}
	case "all":
		for _, item := range common.GetPluginAbleComponents() {
			o.Name = item
			if err = o.uninstallRunE(cmd, args); err != nil {
				return
			}
		}
		return
	default:
		err = fmt.Errorf("not support [%s] yet", o.Name)
		return
	}
	comp := o.getComponent(o.Name)

	_, err = o.Client.Resource(kstypes.GetClusterConfiguration()).Namespace(ns).Patch(ctx,
		name, types.JSONPatchType,
		[]byte(patch),
		metav1.PatchOptions{})
	if err = comp.Uninstall(); err != nil {
		return err
	}

	return
}

func (o *Option) getComponent(name string) Component {
	allComponents := make(map[string]Component)

	alerting := &component.Alerting{}
	allComponents[alerting.GetName()] = alerting

	auditing := &component.Auditing{}
	allComponents[auditing.GetName()] = auditing

	devops := &component.DevOps{Client: o.Client, Clientset: o.Clientset}
	allComponents[devops.GetName()] = devops

	events := &component.Events{}
	allComponents[events.GetName()] = events

	logging := &component.Logging{}
	allComponents[logging.GetName()] = logging

	metricServer := &component.MetricsServer{}
	allComponents[metricServer.GetName()] = metricServer

	networkPolicy := &component.NetworkPolicy{}
	allComponents[networkPolicy.GetName()] = networkPolicy

	openPitrix := &component.OpenPitrix{}
	allComponents[openPitrix.GetName()] = openPitrix

	serviceMesh := &component.ServiceMesh{}
	allComponents[serviceMesh.GetName()] = serviceMesh

	return allComponents[name]
}
