package disableOperation

import "github.com/kubesphere-sigs/ks/kubectl-plugin/common"

type Events struct {
}

func (e *Events) DeleteRelatedResource() error {
	err := common.ExecCommand("helm", "delete", "ks-events", "-n", "kubesphere-logging-system")

	return err
}
