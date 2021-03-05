package component

import (
	"context"
	"fmt"
	"github.com/linuxsuren/ks/kubectl-plugin/common"
	kstypes "github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func newComponentsExecCmd(client dynamic.Interface) (cmd *cobra.Command) {
	availableComs := common.ArrayCompletion("jenkins", "apiserver")

	cmd = &cobra.Command{
		Use:   "exec",
		Short: "Execute a command in a container.",
		Long: `Execute a command in a container.
This command is similar with kubectl exec, the only difference is that you don't need to type the fullname'`,
		ValidArgsFunction: availableComs,
		Args:              cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var kubectl string
			if kubectl, err = exec.LookPath("kubectl"); err != nil {
				return
			}

			switch args[0] {
			case "jenkins":
				var jenkinsPodName string
				var list *unstructured.UnstructuredList
				if list, err = client.Resource(kstypes.GetPodSchema()).Namespace("kubesphere-devops-system").List(
					context.TODO(), metav1.ListOptions{}); err == nil {
					for _, item := range list.Items {
						if strings.HasPrefix(item.GetName(), "ks-jenkins") {
							jenkinsPodName = item.GetName()
						}
					}
				} else {
					fmt.Println(err)
					return
				}

				if jenkinsPodName == "" {
					err = fmt.Errorf("cannot found ks-jenkins pod")
				} else {
					err = syscall.Exec(kubectl, []string{"kubectl", "-n", "kubesphere-devops-system", "exec", "-it", jenkinsPodName, "bash"}, os.Environ())
				}
			case "apiserver":
				var apiserverPodName string
				var list *unstructured.UnstructuredList
				if list, err = client.Resource(kstypes.GetPodSchema()).Namespace("kubesphere-system").List(
					context.TODO(), metav1.ListOptions{}); err == nil {
					for _, item := range list.Items {
						if strings.HasPrefix(item.GetName(), "ks-apiserver") {
							apiserverPodName = item.GetName()
						}
					}
				} else {
					fmt.Println(err)
					return
				}

				if apiserverPodName == "" {
					err = fmt.Errorf("cannot found ks-jenkins pod")
				} else {
					err = syscall.Exec(kubectl, []string{"kubectl", "-n", "kubesphere-system", "exec", "-it", apiserverPodName, "sh"}, os.Environ())
				}
			}
			return
		},
	}

	return
}
