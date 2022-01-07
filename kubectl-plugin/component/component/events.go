package component

import "github.com/kubesphere-sigs/ks/kubectl-plugin/common"

// Events return the struct of Events
type Events struct {
}

// GetName return the name of Events
func (e *Events) GetName() string {
	return "events"
}

// Uninstall uninstall Events
func (e *Events) Uninstall() error {
	err := common.ExecCommand("helm", "delete", "ks-events", "-n", "kubesphere-logging-system")

	return err
}
