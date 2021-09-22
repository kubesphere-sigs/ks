package component

import (
	"context"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	kstypes "github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func newComponentsExecCmd() (cmd *cobra.Command) {
	opt := &Option{}
	cmd = &cobra.Command{
		Use:   "exec",
		Short: "Execute a command in a container.",
		Long: `Execute a command in a container.
This command is similar with kubectl exec, the only difference is that you don't need to type the fullname'`,
		ValidArgsFunction: common.KubeSphereDeploymentCompletion(),
		Args:              cobra.MinimumNArgs(1),
		RunE:              opt.runE,
	}
	return
}

func (o *Option) runE(cmd *cobra.Command, args []string) (err error) {
	ctx := cmd.Root().Context()
	o.Client = common.GetDynamicClient(ctx)
	o.Clientset = common.GetClientset(ctx)

	var kubectl string
	if kubectl, err = exec.LookPath("kubectl"); err != nil {
		return
	}

	var podName string
	var ns string
	if ns, podName, err = o.getPod(args[0]); err == nil {
		err = syscall.Exec(kubectl, []string{"kubectl", "-n", ns, "exec", "-it", podName, "bash"}, os.Environ())
	}
	return
}

func (o *Option) getPod(name string) (ns, podName string, err error) {
	var deployName string
	var list *unstructured.UnstructuredList
	ns, deployName = o.getNsAndName(name)
	if list, err = o.Client.Resource(kstypes.GetPodSchema()).Namespace(ns).List(
		context.TODO(), metav1.ListOptions{}); err == nil {
		for _, item := range list.Items {
			if strings.HasPrefix(item.GetName(), deployName) {
				podName = item.GetName()
				break
			}
		}
	}

	if podName == "" {
		err = fmt.Errorf("cannot found %s pod", deployName)
	}
	return
}
