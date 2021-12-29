package component

import (
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
)

// ServiceMesh return the struct of ServiceMesh
type ServiceMesh struct {
}

// Uninstall uninstall ServiceMesh
func (s *ServiceMesh) Uninstall() error {
	if err := common.ExecCommand("wget", "https://github.com/istio/istio/releases/download/1.12.1/istioctl-1.12.1-linux-amd64.tar.gz"); err != nil {
		return err
	}
	if err := common.ExecCommand("tar", "-xvf", "istioctl-1.12.1-linux-amd64.tar.gz"); err != nil {
		return err
	}
	if err := common.ExecCommand("cp", "istioctl", "/usr/local/bin"); err != nil {
		return err
	}

	if err := common.ExecCommand("istioctl", "x", "uninstall", "--purge"); err != nil {
		return err
	}
	if err := common.ExecCommand("kubectl", "-n", "istio-system", "delete", "kiali", "kiali"); err != nil {
		return err
	}
	if err := common.ExecCommand("helm", "-n", "istio-system", "delete", "kiali-operator"); err != nil {
		return err
	}
	if err := common.ExecCommand("kubectl", "-n", "istio-system", "delete", "jaeger", "jaeger"); err != nil {
		return err
	}
	if err := common.ExecCommand("helm", "-n", "istio-system", "delete", "jaeger-operator"); err != nil {
		return err
	}

	return nil
}
