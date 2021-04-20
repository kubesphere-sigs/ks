package config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/linuxsuren/ks/kubectl-plugin/common"
	kstypes "github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

func newClusterCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := clusterOption{
		client: client,
	}

	cmd = &cobra.Command{
		Use:      "cluster",
		Short:    "Set multi cluster",
		RunE:     opt.runE,
		PostRunE: opt.postRunE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.role, "role", "r", "",
		"Set current KubeSphere cluster as your desired role (none, host, member)")
	flags.StringVarP(&opt.jwtSecret, "jwtSecret", "", "",
		"Need this if you want to set the cluster as the member role")

	_ = cmd.RegisterFlagCompletionFunc("role", common.ArrayCompletion("none", "host", "member"))
	return
}

func (o *clusterOption) runE(cmd *cobra.Command, args []string) (err error) {
	switch o.role {
	case "member":
		if o.jwtSecret == "" {
			err = errors.New("please provide the jwtSecret from the host cluster")
		} else {
			err = o.updateJwtSecret()
		}

		if err != nil {
			return
		}
		fallthrough
	case "none", "host":
		err = o.updateClusterRole()
	case "":
	default:
		err = fmt.Errorf("invalid cluster role: %s", o.role)
	}
	return
}

func (o *clusterOption) postRunE(cmd *cobra.Command, args []string) (err error) {
	err = o.showClusterRole()
	return
}

func (o *clusterOption) updateJwtSecret() (err error) {
	patch := fmt.Sprintf(`[{"op": "replace", "path": "/spec/authentication/jwtSecret", "value": "%s"}]`, o.jwtSecret)
	ctx := context.TODO()
	_, err = o.client.Resource(kstypes.GetClusterConfiguration()).Namespace("kubesphere-system").Patch(ctx,
		"ks-installer", types.JSONPatchType,
		[]byte(patch),
		metav1.PatchOptions{})
	return
}

func (o *clusterOption) updateClusterRole() (err error) {
	patch := fmt.Sprintf(`[{"op": "replace", "path": "/spec/multicluster/clusterRole", "value": "%s"}]`, o.role)
	ctx := context.TODO()
	_, err = o.client.Resource(kstypes.GetClusterConfiguration()).Namespace("kubesphere-system").Patch(ctx,
		"ks-installer", types.JSONPatchType,
		[]byte(patch),
		metav1.PatchOptions{})
	return
}

func (o *clusterOption) showClusterRole() (err error) {
	ctx := context.TODO()
	var rawData *unstructured.Unstructured
	if rawData, err = o.client.Resource(kstypes.GetClusterConfiguration()).Namespace("kubesphere-system").
		Get(ctx, "ks-installer", metav1.GetOptions{}); err == nil {
		var data []byte
		buf := bytes.NewBuffer(data)
		enc := json.NewEncoder(buf)
		if err = enc.Encode(rawData); err != nil {
			return
		}

		var yamlData []byte
		if yamlData, err = yaml.JSONToYAML(buf.Bytes()); err != nil {
			return
		}

		installer := common.KSInstaller{}
		if err = yaml.Unmarshal(yamlData, &installer); err == nil {
			fmt.Printf("cluster role: %s\n", installer.Spec.Multicluster.ClusterRole)
			if installer.Spec.Multicluster.ClusterRole == "member" {
				fmt.Printf("host cluster jwtSecret: %s\n", installer.Spec.Authentication.JwtSecret)
			}
		}
	}

	var rawConfigMap *unstructured.Unstructured
	if rawConfigMap, err = o.client.Resource(kstypes.GetConfigMapSchema()).Namespace("kubesphere-system").
		Get(context.TODO(), "kubesphere-config", metav1.GetOptions{}); err == nil {
		data := rawConfigMap.Object["data"]
		dataMap := data.(map[string]interface{})
		kubeSphereCfg := dataMap["kubesphere.yaml"]

		cfg := kubeSphereConfig{}
		if err = yaml.Unmarshal([]byte(fmt.Sprintf("%v", kubeSphereCfg)), &cfg); err == nil {
			fmt.Printf("current jwtSecret: %s\n", cfg.Authentication.JwtSecret)
		}
	}
	return
}
