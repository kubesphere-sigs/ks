package auth

import (
	"context"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"strings"
)

func authAddCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &authAddOption{
		Client: client,
	}

	cmd = &cobra.Command{
		Use:     "add",
		Short:   "Add addition auth configuration into kubesphere-config",
		PreRunE: opt.preRunE,
		Example: `
subjects:
- apiGroup: users.iam.kubesphere.io
  kind: Group
  name: pre-registration
`,
		RunE: opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.Type, "type", "t", "",
		"The oAuth provider, supported: GitHub, Aliyun, Gitee")
	flags.StringVarP(&opt.ClientID, "client-id", "", "",
		"The client id which you can find it from the oAuth provider")
	flags.StringVarP(&opt.ClientSecret, "client-secret", "", "",
		"The client secret which you can find it from the oAuth provider")
	flags.StringVarP(&opt.Host, "host", "", "",
		"The host of KubeSphere")
	return
}

type authAddOption struct {
	Client dynamic.Interface

	Type string

	ClientID     string
	ClientSecret string
	Host         string
}

func (o *authAddOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	if o.ClientID == "" || o.ClientSecret == "" || o.Host == "" {
		return fmt.Errorf("ClientID, ClientSecret, Host cannot be empty")
	}

	// make sure the host has prefix http or https
	if !strings.HasPrefix(o.Host, "http://") && !strings.HasPrefix(o.Host, "https://") {
		o.Host = fmt.Sprintf("http://%s", o.Host)
	}

	switch o.Type {
	case "GitHub", "Aliyun", "Gitee":
	default:
		err = fmt.Errorf("not support auth type: %s", o.Type)
	}
	return
}

func (o *authAddOption) runE(cmd *cobra.Command, args []string) (err error) {
	var authConfig string
	switch o.Type {
	case "GitHub":
		authConfig = getGitHubAuth(*o)
	case "Aliyun":
		authConfig = getAliyunAuth(*o)
	case "Gitee":
		authConfig = getGiteeAuth(*o)
	}

	var rawConfigMap *unstructured.Unstructured
	if rawConfigMap, err = o.Client.Resource(types.GetConfigMapSchema()).Namespace("kubesphere-system").
		Get(context.TODO(), "kubesphere-config", metav1.GetOptions{}); err == nil {
		data := rawConfigMap.Object["data"]
		dataMap := data.(map[string]interface{})
		result := updateAuthentication(dataMap["kubesphere.yaml"].(string), o.Type, authConfig)
		if strings.TrimSpace(result) == "" {
			err = fmt.Errorf("error happend when parse kubesphere-config")
			return
		}

		rawConfigMap.Object["data"] = map[string]interface{}{
			"kubesphere.yaml": result,
		}
		_, err = o.Client.Resource(types.GetConfigMapSchema()).Namespace("kubesphere-system").Update(context.TODO(),
			rawConfigMap, metav1.UpdateOptions{})
	} else {
		err = fmt.Errorf("cannot found configmap kubesphere-config, %v", err)
	}

	return
}
