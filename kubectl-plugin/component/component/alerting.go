package component

import "github.com/kubesphere-sigs/ks/kubectl-plugin/common"

// Alerting return the struct of Alerting
type Alerting struct {
}

// GetName return the name of Alerting
func (e *Alerting) GetName() string {
	return "alerting"
}

// Uninstall uninstall Alerting
func (e *Alerting) Uninstall() error {
	err := common.ExecCommand("kubectl", "-n", "kubesphere-monitoring-system", "delete", "thanosruler", "kubesphere")

	return err
}
