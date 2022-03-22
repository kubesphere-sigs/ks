package component

import "github.com/kubesphere-sigs/ks/utils/helm"

// Events return the struct of Events
type Events struct {
}

// GetName return the name of Events
func (e *Events) GetName() string {
	return "events"
}

// Uninstall uninstall Events
func (e *Events) Uninstall() error {
	uninstallRequest := helm.UninstallRequest{
		ComponentName: "ks-events",
		Namespace:     "kubesphere-logging-system",
		KubeConfig:    "/root/.kube/config",
	}
	err := uninstallRequest.Do()

	return err
}
