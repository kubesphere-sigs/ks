package component

import "github.com/kubesphere-sigs/ks/kubectl-plugin/common"

// ServiceMesh return the struct of ServiceMesh
type ServiceMesh struct {
}

// Uninstall uninstall ServiceMesh
func (s *ServiceMesh) Uninstall() error {
	_ = common.ExecCommand("curl", "-L", "https://istio.io/downloadIstio")
	_ = common.ExecCommand("sh", "downloadIstioCandidate.sh")

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
