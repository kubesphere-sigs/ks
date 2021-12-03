package disableOperation

import "github.com/kubesphere-sigs/ks/kubectl-plugin/common"

type Auditing struct {
}

func (e *Auditing) DeleteRelatedResource() error {
	if err := common.ExecCommand("helm", "uninstall", "kube-auditing", "-n", "kubesphere-monitoring-system"); err != nil {
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
