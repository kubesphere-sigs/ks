package component

import "github.com/kubesphere-sigs/ks/kubectl-plugin/common"

// Alerting return the struct of Alerting
type Alerting struct {
}

func (e *Alerting) Uninstall() error {
	err := common.ExecCommand("kubectl", "-n", "kubesphere-monitoring-system", "delete", "thanosruler", "kubesphere")

	return err
}
