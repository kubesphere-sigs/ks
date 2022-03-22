package component

import (
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/utils/helm"
	"github.com/linuxsuren/http-downloader/pkg/installer"
)

// ServiceMesh return the struct of ServiceMesh
type ServiceMesh struct {
}

// GetName return the name of ServiceMesh
func (s *ServiceMesh) GetName() string {
	return "servicemesh"
}

// Uninstall uninstall ServiceMesh
func (s *ServiceMesh) Uninstall() error {
	is := installer.Installer{
		Provider: "github",
	}
	if err := is.CheckDepAndInstall(map[string]string{
		"istioctl": "istio/istio",
	}); err != nil {
		return err
	}

	if err := common.ExecCommand("istioctl", "x", "uninstall", "--purge"); err != nil {
		return err
	}
	if err := common.ExecCommand("kubectl", "-n", "istio-system", "delete", "kiali", "kiali"); err != nil {
		return err
	}
	uninstallRequest := helm.UninstallRequest{
		ComponentName: "kiali-operator",
		Namespace:     "istio-system",
		KubeConfig:    "/root/.kube/config",
	}
	if err := uninstallRequest.Do(); err != nil {
		return err
	}
	if err := common.ExecCommand("kubectl", "-n", "istio-system", "delete", "jaeger", "jaeger"); err != nil {
		return err
	}
	uninstallRequest = helm.UninstallRequest{
		ComponentName: "jaeger-operator",
		Namespace:     "istio-system",
		KubeConfig:    "/root/.kube/config",
	}
	if err := uninstallRequest.Do(); err != nil {
		return err
	}
	return nil
}
