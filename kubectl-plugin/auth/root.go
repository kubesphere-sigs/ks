package auth

import (
	"context"
	"fmt"
	"github.com/linuxsuren/ks/kubectl-plugin/common"
	"github.com/linuxsuren/ks/kubectl-plugin/types"
	kstypes "github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"strings"
)

type authdOption struct {
	multipleLogin string
	kubectlImage  string

	client dynamic.Interface
}

// NewAuthCmd returns a command of auth
func NewAuthCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &authdOption{
		client: client,
	}

	cmd = &cobra.Command{
		Use:   "auth",
		Short: "Configure auth for KubeSphere",
		RunE:  opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.multipleLogin, "multipleLogin", "", "",
		"Enable or disable multipleLogin with value: enable, disable")
	flags.StringVarP(&opt.kubectlImage, "kubectlImage", "", "",
		"Update the kubectl image in KubeSphere. Available value could be: ks, default, custom:your-image. The default image is kubesphere/kubectl:v1.17.0")

	_ = cmd.RegisterFlagCompletionFunc("multipleLogin", common.ArrayCompletion("enable", "disable"))
	_ = cmd.RegisterFlagCompletionFunc("kubectlImage", common.ArrayCompletion("ks", "default", "custom:"))
	cmd.AddCommand(authAddCmd(client))
	return
}

func (o *authdOption) runE(cmd *cobra.Command, args []string) (err error) {
	if err = o.updateMultipleLogin(); err != nil {
		return
	}

	if err = o.updateKubectlImage(); err != nil {
		return
	}
	return
}

func (o *authdOption) updateMultipleLogin() (err error) {
	var enable bool
	switch o.multipleLogin {
	case "enable":
		enable = true
	case "disable":
		enable = false
	default:
		return
	}

	ctx := context.TODO()
	var rawConfigMap *unstructured.Unstructured
	if rawConfigMap, err = o.client.Resource(types.GetConfigMapSchema()).Namespace("kubesphere-system").
		Get(ctx, "kubesphere-config", metav1.GetOptions{}); err == nil {
		data := rawConfigMap.Object["data"]
		dataMap := data.(map[string]interface{})
		result := setMultipleLogin(dataMap["kubesphere.yaml"].(string), enable)
		if strings.TrimSpace(result) == "" {
			err = fmt.Errorf("error happend when parse kubesphere-config")
			return
		}

		rawConfigMap.Object["data"] = map[string]interface{}{
			"kubesphere.yaml": result,
		}
		_, err = o.client.Resource(types.GetConfigMapSchema()).Namespace("kubesphere-system").Update(ctx,
			rawConfigMap, metav1.UpdateOptions{})
	}

	if err == nil {
		err = o.killAPIServer()
	}
	return
}

func (o *authdOption) updateKubectlImage() (err error) {
	var image string
	switch o.kubectlImage {
	case "":
		return
	case "ks":
		image = "surenpi/ks:v0.0.26"
	case "default":
		image = "kubesphere/kubectl:v1.17.0"
	default:
		if strings.HasPrefix(o.kubectlImage, "custom:") {
			image = strings.ReplaceAll(o.kubectlImage, "custom:", "")

			if image == "" {
				err = fmt.Errorf("custom image is empty")
				return
			}
		}
	}

	ctx := context.TODO()
	var rawConfigMap *unstructured.Unstructured
	if rawConfigMap, err = o.client.Resource(types.GetConfigMapSchema()).Namespace("kubesphere-system").
		Get(ctx, "kubesphere-config", metav1.GetOptions{}); err == nil {
		data := rawConfigMap.Object["data"]
		dataMap := data.(map[string]interface{})
		result := setKubectlImage(dataMap["kubesphere.yaml"].(string), image)
		if strings.TrimSpace(result) == "" {
			err = fmt.Errorf("error happend when parse kubesphere-config")
			return
		}

		rawConfigMap.Object["data"] = map[string]interface{}{
			"kubesphere.yaml": result,
		}
		_, err = o.client.Resource(types.GetConfigMapSchema()).Namespace("kubesphere-system").Update(ctx,
			rawConfigMap, metav1.UpdateOptions{})
	}

	if err == nil {
		err = o.killAPIServer()
	}
	return
}

func (o *authdOption) killAPIServer() (err error) {
	ctx := context.TODO()

	err = o.client.Resource(kstypes.GetPodSchema()).Namespace("kubesphere-system").DeleteCollection(ctx, metav1.DeleteOptions{
		TypeMeta: metav1.TypeMeta{
			Kind: "pod",
		},
	}, metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": "ks-apiserver",
		}).String(),
	})
	return
}
