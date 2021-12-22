package component

import "github.com/kubesphere-sigs/ks/kubectl-plugin/common"

// Events return the struct of Events
type Events struct {
}

func (e *Events) Uninstall() error {
	err := common.ExecCommand("helm", "delete", "ks-events", "-n", "kubesphere-logging-system")

	return err
}
