package component

import (
	"context"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/component/disableOperation"
	kstypes "github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
)

// EnableOption is the option for component enable command
type EnableOption struct {
	Option

	Edit   bool
	Toggle bool
}

type DisableOperation interface {
	DeleteRelatedResource() error
}

// newComponentEnableCmd returns a command to enable (or disable) a component by name
func newComponentEnableCmd() (cmd *cobra.Command) {
	opt := &EnableOption{}
	cmd = &cobra.Command{
		Use:   "enable",
		Short: "Enable or disable the specific KubeSphere component",
		Example: `You can enable a single component with name via: ks com enable devops
Or it's possible to enable all components via: ks com enable all`,
		PreRunE:           opt.enablePreRunE,
		ValidArgsFunction: common.PluginAbleComponentsCompletion(),
		RunE:              opt.enableRunE,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opt.Edit, "edit", "e", false,
		"Indicate if you want to edit it instead of enable/disable a specified one. This flag will make others not work.")
	flags.BoolVarP(&opt.Toggle, "toggle", "t", false,
		"Indicate if you want to disable a component")
	flags.StringVarP(&opt.Name, "name", "n", "",
		"The name of target component which you want to enable/disable. Please provide option --sonarqube if you want to enable SonarQube.")
	flags.StringVarP(&opt.SonarQube, "sonarqube", "", "",
		"The SonarQube URL")
	flags.StringVarP(&opt.SonarQube, "sonar", "", "",
		"The SonarQube URL")
	flags.StringVarP(&opt.SonarQubeToken, "sonarqube-token", "", "",
		"The token of SonarQube")

	_ = cmd.RegisterFlagCompletionFunc("name", common.PluginAbleComponentsCompletion())

	// these are aliased options
	_ = flags.MarkHidden("sonar")
	return
}

func (o *EnableOption) enablePreRunE(cmd *cobra.Command, args []string) (err error) {
	ctx := cmd.Root().Context()
	o.Client = common.GetDynamicClient(ctx)
	o.Clientset = common.GetClientset(ctx)

	if o.Edit {
		return
	}

	return o.componentNameCheck(cmd, args)
}

func (o *EnableOption) enableRunE(cmd *cobra.Command, args []string) (err error) {
	ctx := context.TODO()
	if o.Edit {
		err = common.UpdateWithEditor(kstypes.GetClusterConfiguration(), "kubesphere-system", "ks-installer", o.Client)
	} else {
		enabled := strconv.FormatBool(!o.Toggle)
		ns, name := "kubesphere-system", "ks-installer"
		var patchTarget string
		switch o.Name {
		case "devops", "alerting", "auditing", "events", "logging", "metrics_server", "networkpolicy", "notification", "openpitrix", "servicemesh":
			patchTarget = o.Name
		case "sonarqube", "sonar":
			if o.SonarQube == "" || o.SonarQubeToken == "" {
				err = fmt.Errorf("SonarQube or token is empty, please provide --sonarqube")
			} else {
				name = "ks-console-config"
				err = integrateSonarQube(o.Client, ns, name, o.SonarQube, o.SonarQubeToken)
			}
			return
		case "metering":
			patchTarget = "metering"
			if _, err = o.Client.Resource(kstypes.GetConfigMapSchema()).Namespace("kubesphere-system").
				Get(ctx, "ks-metering-config", metav1.GetOptions{}); err != nil {
				var data *unstructured.Unstructured
				if data, err = kstypes.GetObjectFromYaml(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: ks-metering-config
data:
  ks-metering.yaml: |
    retentionDay: 7d
    billing:
      priceInfo:
        currencyUnit: "USD"
        cpuPerCorePerHour: 1.5
        memPerGigabytesPerHour: 5
        ingressNetworkTrafficPerGiagabytesPerHour: 3.5
        egressNetworkTrafficPerGigabytesPerHour: 0.5
        pvcPerGigabytesPerHour: 2.1`); err == nil {
					_, err = o.Client.Resource(kstypes.GetConfigMapSchema()).Namespace("kubesphere-system").Create(ctx, data, metav1.CreateOptions{})
				}
			}
		case "all":
			for _, item := range common.GetPluginAbleComponents() {
				o.Name = item
				if err = o.enableRunE(cmd, args); err != nil {
					return
				}
			}
			return
		default:
			err = fmt.Errorf("not support [%s] yet", o.Name)
			return
		}

		if err == nil {
			patch := fmt.Sprintf(`[{"op": "replace", "path": "/spec/%s/enabled", "value": %s}]`, patchTarget, enabled)
			_, err = o.Client.Resource(kstypes.GetClusterConfiguration()).Namespace(ns).Patch(ctx,
				name, types.JSONPatchType,
				[]byte(patch),
				metav1.PatchOptions{})

			ops := disableOperation.DevOps{}
			err = ops.DeleteRelatedResource()
		}
	}
	return
}
