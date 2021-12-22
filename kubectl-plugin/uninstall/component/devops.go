package component

import (
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"strings"
)

// DevOps return the struct of DevOps
type DevOps struct {
}

func (o *DevOps) Uninstall() error {
	// Uninstall DevOps application
	if err := common.ExecCommand("helm", "uninstall", "-n", "kubesphere-devops-system", "devops"); err != nil {
		return err
	}

	// Remove DevOps installation status
	patch := fmt.Sprintf(`'[{"op": "remove", "path": "/status/devops"}]'`)
	if err := common.ExecCommand("kubectl", "patch", "-n", "kubesphere-system", "cc", "ks-installer", "--type=json", "-p", patch); err != nil {
		return err
	}

	// ############# DevOps Resource Deletion ##############
	// Remove all resources related with DevOps
	crdStr, err := common.ExecCommandGetOutput("kubectl", "get", "crd", fmt.Sprintf(`-o=jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}'`))
	if err != nil {
		return err
	}
	nsStr, err := common.ExecCommandGetOutput("kubectl", "get", "ns", fmt.Sprintf(`-ojsonpath='{.items..metadata.name}'`))
	if err != nil {
		return err
	}
	for _, devopsCrd := range strings.Split(crdStr, "\n") {
		if !strings.Contains(devopsCrd, "devops.kubesphere.io") {
			continue
		}
		for _, ns := range strings.Split(nsStr, " ") {
			devopsResStr, _ := common.ExecCommandGetOutput("kubectl", "get", devopsCrd, "-n", ns, "-oname")
			for _, devopsRes := range strings.Split(devopsResStr, "\n") {
				if err = common.ExecCommand("kubectl", "patch", devopsRes, "-n", ns, "-p", fmt.Sprintf(`'{"metadata":{"finalizers":[]}}'`), "--type=merge"); err != nil {
					return err
				}
			}
		}
		// Remove all DevOps CRDs
		if err = common.ExecCommand("kubectl", "delete", "crd", devopsCrd); err != nil {
			return err
		}

	}

	// Remove DevOps namespace
	if err = common.ExecCommand("kubectl", "delete", "namespace", "kubesphere-devops-system"); err != nil {
		return err
	}

	return nil
}
