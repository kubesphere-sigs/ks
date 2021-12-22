package component

import "github.com/kubesphere-sigs/ks/kubectl-plugin/common"

// MetricsServer return the struct of MetricsServer
type MetricsServer struct {
}

// Uninstall uninstall MetricsServer
func (m *MetricsServer) Uninstall() error {
	if err := common.ExecCommand("kubectl", "delete", "apiservice", "v1beta1.metrics.k8s.io"); err != nil {
		return err
	}
	if err := common.ExecCommand("kubectl", "-n", "kube-system", "delete", "service", "metrics-server"); err != nil {
		return err
	}
	if err := common.ExecCommand("kubectl", "-n", "kube-system", "delete", "deployment", "metrics-server"); err != nil {
		return err
	}

	return nil
}
