package component

import (
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/utils/helm"
)

// Auditing return the struct of Auditing
type Auditing struct {
}

// GetName return the name of Auditing
func (e *Auditing) GetName() string {
	return "auditing"
}

// Uninstall uninstall Auditing
func (e *Auditing) Uninstall() error {
	uninstallRequest := helm.UninstallRequest{
		ComponentName: "kube-auditing",
		Namespace:     "kubesphere-monitoring-system",
		KubeConfig:    "/root/.kube/config",
	}
	if err := uninstallRequest.Do(); err != nil {
		return err
	}
	if err := common.ExecCommand("kubectl", "delete", "crd", "awh"); err != nil {
		return err
	}
	if err := common.ExecCommand("kubectl", "delete", "crd", "ar"); err != nil {
		return err
	}

	return nil
}
