package disableOperation

import "github.com/kubesphere-sigs/ks/kubectl-plugin/common"

type ServiceMesh struct {
}

func (s *ServiceMesh) DeleteRelatedResource() error {
	_ = common.ExecCommand("curl", "-L", "https://istio.io/downloadIstio", "|", "sh", "-")
	_ = common.ExecCommand("istioctl", "x", "uninstall", "--purge")
	_ = common.ExecCommand("kubectl", "-n", "istio-system", "delete", "kiali", "kiali")
	_ = common.ExecCommand("helm", "-n", "istio-system", "delete", "kiali-operator")
	_ = common.ExecCommand("kubectl", "-n", "istio-system", "delete", "jaeger", "jaeger")
	_ = common.ExecCommand("helm", "-n", "istio-system", "delete", "jaeger-operator")

	return nil
}
