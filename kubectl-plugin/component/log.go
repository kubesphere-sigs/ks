package component

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	kstypes "github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
)

// LogOption is the option for component log command
type LogOption struct {
	Option

	Follow bool
	Tail   int64
}

// newComponentLogCmd returns a command to enable (or disable) a component by name
func newComponentLogCmd() (cmd *cobra.Command) {
	opt := &LogOption{}
	cmd = &cobra.Command{
		Use:               "log",
		Short:             "Output the log of KubeSphere component",
		ValidArgsFunction: common.KubeSphereDeploymentCompletion(),
		PreRunE:           opt.componentNameCheck,
		RunE:              opt.logRunE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.Name, "name", "n", "",
		"The name of target component which you want to reset.")
	flags.BoolVarP(&opt.Follow, "follow", "f", true,
		"Specify if the logs should be streamed.")
	flags.Int64VarP(&opt.Tail, "tail", "", 50,
		`Lines of recent log file to display.`)
	return
}

func (o *LogOption) logRunE(cmd *cobra.Command, args []string) (err error) {
	if o.Clientset == nil {
		err = fmt.Errorf("kubernetes clientset is nil")
		return
	}

	ctx := context.TODO()
	var ns, name string
	if ns, name = o.getNsAndName(o.Name); name == "" {
		err = fmt.Errorf("not supported yet: %s", o.Name)
		return
	}

	var data []byte
	buf := bytes.NewBuffer(data)
	var rawPip *unstructured.Unstructured
	deploy := &simpleDeploy{}
	if rawPip, err = o.Client.Resource(kstypes.GetDeploySchema()).Namespace(ns).Get(ctx, name, metav1.GetOptions{}); err == nil {
		enc := json.NewEncoder(buf)
		enc.SetIndent("", "    ")
		if err = enc.Encode(rawPip); err != nil {
			return
		}

		if err = json.Unmarshal(buf.Bytes(), deploy); err != nil {
			return
		}
	}

	var podList *v1.PodList
	var podName string
	if podList, err = o.Clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(deploy.Spec.Selector.MatchLabels).String(),
	}); err == nil {
		if len(podList.Items) > 0 {
			podName = podList.Items[0].Name
		}
	} else {
		return
	}

	if podName == "" {
		err = fmt.Errorf("cannot found the pod with deployment '%s'", name)
		return
	}

	if len(deploy.Spec.Selector.MatchLabels) > 0 {
		req := o.Clientset.CoreV1().Pods(ns).GetLogs(podName, &v1.PodLogOptions{
			Follow:    o.Follow,
			TailLines: &o.Tail,
		})
		var podLogs io.ReadCloser
		if podLogs, err = req.Stream(context.TODO()); err == nil {
			defer func() {
				_ = podLogs.Close()
			}()

			_, err = io.Copy(cmd.OutOrStdout(), podLogs)
		}
	}
	return
}
