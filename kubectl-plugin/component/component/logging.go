package component

import (
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/utils/helm"
)

// Logging return the struct of Logging
type Logging struct{}

// GetName return the name of Logging
func (l *Logging) GetName() string {
	return "logging"
}

// Uninstall uninstall Logging
func (l *Logging) Uninstall() error {
	// To disable only log collection
	if err := common.ExecCommand("kubectl", "delete", "inputs.logging.kubesphere.io", "-n", "kubesphere-logging-system", "tail"); err != nil {
		return err
	}

	// To uninstall Logging system including Elasticsearch
	if err := common.ExecCommand("kubectl", "delete", "crd", "fluentbitconfigs.logging.kubesphere.io"); err != nil {
		return err
	}
	if err := common.ExecCommand("kubectl", "delete", "crd", "fluentbits.logging.kubesphere.io"); err != nil {
		return err
	}
	if err := common.ExecCommand("kubectl", "delete", "crd", "inputs.logging.kubesphere.io"); err != nil {
		return err
	}
	if err := common.ExecCommand("kubectl", "delete", "crd", "outputs.logging.kubesphere.io"); err != nil {
		return err
	}
	if err := common.ExecCommand("kubectl", "delete", "crd", "parsers.logging.kubesphere.io"); err != nil {
		return err
	}
	if err := common.ExecCommand("kubectl", "delete", "deployments.apps", "-n", "kubesphere-logging-system", "fluentbit-operator"); err != nil {
		return err
	}

	uninstallRequest := helm.UninstallRequest{
		ComponentName: "elasticsearch-logging",
		Namespace:     "kubesphere-logging-system",
		KubeConfig:    "/root/.kube/config",
	}
	if err := uninstallRequest.Do(); err != nil {
		return err
	}

	return nil
}
