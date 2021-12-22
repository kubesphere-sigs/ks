package component

import (
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
)

type DevOps struct {
}

func (o *DevOps) Uninstall() error {
	// Uninstall DevOps application
	if err := common.ExecCommand("helm", "uninstall", "-n", "kubesphere-devops-system", "devops"); err != nil {
		return err
	}

	// Remove DevOps installation status
	patch := `-p='[{"op": "remove", "path": "/status/devops"}]'`
	if err := common.ExecCommand("kubectl", "patch", "-n", "kubesphere-system", "cc", "ks-installer", "--type=json", patch); err != nil {
		return err
	}
	patch = `-p='[{"op": "replace", "path": "/spec/devops/enabled", "value": false}]'`
	if err := common.ExecCommand("kubectl", "patch", "-n", "kubesphere-system", "cc", "ks-installer", "--type=json", patch); err != nil {
		return err
	}

	// ############# DevOps Resource Deletion ##############
	// Remove all resources related with DevOps
	arg := `for devops_crd in $(kubectl get crd -o=jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' | grep "devops.kubesphere.io"); do
    for ns in $(kubectl get ns -ojsonpath='{.items..metadata.name}'); do
        for devops_res in $(kubectl get $devops_crd -n $ns -oname); do
            kubectl patch $devops_res -n $ns -p '{"metadata":{"finalizers":[]}}' --type=merge
        done
    done
done`
	if err := common.ExecCommand("", arg); err != nil {
		return err
	}

	// Remove all DevOps CRDs
	if err := common.ExecCommand("kubectl", "get", "crd", `-o=jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' | grep "devops.kubesphere.io" | xargs -I crd_name kubectl delete crd crd_name`); err != nil {
		return err
	}

	// Remove DevOps namespace
	if err := common.ExecCommand("kubectl", "delete", "namespace", "kubesphere-devops-system"); err != nil {
		return err
	}

	return nil
}
