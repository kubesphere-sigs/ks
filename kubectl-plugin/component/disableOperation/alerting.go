package disableOperation

import "github.com/kubesphere-sigs/ks/kubectl-plugin/common"

type Alerting struct {
}

func (e *Alerting) DeleteRelatedResource() error {
	err := common.ExecCommand("kubectl", "delete", "thanosruler", "kubesphere", "-n", "kubesphere-monitoring-system")

	return err
}
