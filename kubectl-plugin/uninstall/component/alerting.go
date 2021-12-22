package component

import "github.com/kubesphere-sigs/ks/kubectl-plugin/common"

type Alerting struct {
}

func (e *Alerting) Uninstall() error {
	err := common.ExecCommand("kubectl", "-n", "kubesphere-monitoring-system", "delete", "thanosruler", "kubesphere")

	return err
}
