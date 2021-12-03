package disableOperation

import (
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
)

type Logging struct{}

func (l *Logging) DeleteRelatedResource() error {
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
	if err := common.ExecCommand("helm", "uninstall", "elasticsearch-logging", "--namespace", "kubesphere-logging-system"); err != nil {
		return err
	}

	return nil
}
