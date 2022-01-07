package component

import (
	"context"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	kstypes "github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"strings"
)

// DevOps return the struct of DevOps
type DevOps struct {
	Client    dynamic.Interface
	Clientset *kubernetes.Clientset
}

// GetName return the name of DevOps
func (o *DevOps) GetName() string {
	return "devops"
}

// Uninstall uninstall DevOps
func (o *DevOps) Uninstall() error {
	// Uninstall DevOps application
	_ = common.ExecCommand("helm", "uninstall", "-n", "kubesphere-devops-system", "devops")

	// Remove DevOps installation status
	ctx := context.TODO()
	patch := fmt.Sprintf(`[{"op": "remove", "path": "/status/devops"}]`)
	_, err := o.Client.Resource(kstypes.GetClusterConfiguration()).Namespace("kubesphere-system").Patch(ctx,
		"ks-installer", types.JSONPatchType,
		[]byte(patch),
		metav1.PatchOptions{})
	if err != nil {
		return err
	}

	// ############# DevOps Resource Deletion ##############
	// Remove all resources related with DevOps
	crdStr, err := common.ExecCommandGetOutput("kubectl", "get", "crd", fmt.Sprintf(`-o=jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}'`))
	if err != nil {
		return err
	}
	nsStr, err := common.ExecCommandGetOutput("kubectl", "get", "ns", fmt.Sprintf(`-ojsonpath='{.items..metadata.name}'`))
	if err != nil {
		return err
	}
	for _, devopsCrd := range strings.Split(crdStr, "\n") {
		if !strings.Contains(devopsCrd, "devops.kubesphere.io") {
			continue
		}
		for _, ns := range strings.Split(nsStr, " ") {
			ns = strings.Replace(ns, "'", "", -1)
			if len(ns) == 0 {
				continue
			}
			devopsResStr, _ := common.ExecCommandGetOutput("kubectl", "get", devopsCrd, "-n", ns, "-oname")
			for _, devopsRes := range strings.Split(devopsResStr, "\n") {
				if len(devopsRes) == 0 {
					continue
				}
				patch = fmt.Sprintf(`{"metadata":{"finalizers":[]}}`)
				_, err = o.Client.Resource(kstypes.GetClusterConfiguration()).Namespace(ns).Patch(ctx,
					devopsRes, types.JSONPatchType,
					[]byte(patch),
					metav1.PatchOptions{})
				if err != nil && !strings.Contains(err.Error(), "not found") {
					return err
				}
			}
		}
		// Remove all DevOps CRDs
		if err = common.ExecCommand("kubectl", "delete", "crd", devopsCrd); err != nil {
			return err
		}

	}

	// Remove DevOps namespace
	if err = common.ExecCommand("kubectl", "delete", "namespace", "kubesphere-devops-system"); err != nil {
		return err
	}

	return nil
}
