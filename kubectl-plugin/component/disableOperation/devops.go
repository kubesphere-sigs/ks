package disableOperation

import (
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
)

type DevOps struct {
}

func (o *DevOps) DeleteRelatedResource() error {
	// Uninstall DevOps application
	err := common.ExecCommand("helm", "uninstall", "-n", "kubesphere-devops-system", "devops")

	// Remove DevOps installation status
	patch := `-p=[{"op": "remove", "path": "/status/devops"}]`
	err = common.ExecCommand("kubectl", "patch", "-n", "kubesphere-system", "cc", "ks-installer", "--type=json", patch)

	// ############# DevOps Resource Deletion ##############
	// Remove all resources related with DevOps
	arg := `for devops_crd in $(kubectl get crd -o=jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' | grep "devops.kubesphere.io"); do
    for ns in $(kubectl get ns -ojsonpath='{.items..metadata.name}'); do
        for devops_res in $(kubectl get $devops_crd -n $ns -oname); do
            kubectl patch $devops_res -n $ns -p '{"metadata":{"finalizers":[]}}' --type=merge
        done
    done
done`
	err = common.ExecCommand("", arg)

	// Remove unused namespaces
	err = common.ExecCommand("kubectl", "delete", "namespace", "kubesphere-devops-system")

	return err
}
