package config

import (
	"context"
	"errors"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"strings"
)

func newMigrateCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := migrateOption{
		client: client,
	}

	cmd = &cobra.Command{
		Use:     "migrate",
		Short:   "Migrate DevOps into a separate one",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opt.devops, "devops", "", true,
		"Migrate DevOps")
	flags.StringVarP(&opt.service, "service", "", "",
		"The service address of DevOps")
	flags.StringVarP(&opt.namespace, "namespace", "", "kubesphere-devops-system",
		"The namespace of DevOps")
	return
}

func (o *migrateOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	if o.service == "" {
		err = errors.New("the flag --service cannot be empty")
		return
	}
	if o.namespace == "" {
		err = errors.New("the flag --namespace cannot be empty")
	}
	return
}

func (o *migrateOption) runE(cmd *cobra.Command, args []string) (err error) {
	if err = o.updateKubeSphereConfig("kubesphere-config", "kubesphere-system", map[string]interface{}{
		"enable":               false,
		"devopsServiceAddress": o.service,
	}); err != nil {
		return
	}

	var password string
	if password, err = o.getDevOpsPassword(); password == "" {
		err = fmt.Errorf("the password of Jenkins is empty")
	} else if err == nil {
		err = o.updateKubeSphereConfig("devops-config", o.namespace, map[string]interface{}{
			"password": password,
		})
	}
	return
}

func (o *migrateOption) updateKubeSphereConfig(name, namespace string, ksdataMap map[string]interface{}) (err error) {
	var rawConfigMap *unstructured.Unstructured
	if rawConfigMap, err = o.client.Resource(types.GetConfigMapSchema()).Namespace(namespace).
		Get(context.TODO(), name, metav1.GetOptions{}); err == nil {
		data := rawConfigMap.Object["data"]
		dataMap := data.(map[string]interface{})

		result := updateAuthWithObj(dataMap["kubesphere.yaml"].(string), ksdataMap)
		if strings.TrimSpace(result) == "" {
			err = fmt.Errorf("error happend when parse kubesphere-config")
			return
		}

		rawConfigMap.Object["data"] = map[string]interface{}{
			"kubesphere.yaml": result,
		}
		_, err = o.client.Resource(types.GetConfigMapSchema()).Namespace(namespace).Update(context.TODO(),
			rawConfigMap, metav1.UpdateOptions{})
	} else {
		err = fmt.Errorf("cannot found configmap kubesphere-config, %v", err)
	}
	return
}

func (o *migrateOption) getDevOpsPassword() (password string, err error) {
	var rawConfigMap *unstructured.Unstructured
	if rawConfigMap, err = o.client.Resource(types.GetConfigMapSchema()).Namespace("kubesphere-system").
		Get(context.TODO(), "kubesphere-config", metav1.GetOptions{}); err == nil {
		data := rawConfigMap.Object["data"]
		dataMap := data.(map[string]interface{})
		mapData := make(map[string]interface{})
		if err := yaml.Unmarshal([]byte(dataMap["kubesphere.yaml"].(string)), mapData); err == nil {
			var obj interface{}
			var ok bool
			var mapObj map[string]interface{}
			if obj, ok = mapData["devops"]; ok {
				mapObj = obj.(map[string]interface{})
				if passwdObj := mapObj["password"]; passwdObj != nil {
					password = passwdObj.(string)
				}
			}
		}
	} else {
		err = fmt.Errorf("cannot found configmap kubesphere-config, %v", err)
	}
	return
}

func updateAuthWithObj(yamlf string, dataMap map[string]interface{}) string {
	mapData := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(yamlf), mapData); err == nil {
		var obj interface{}
		var ok bool
		var mapObj map[string]interface{}
		if obj, ok = mapData["devops"]; ok {
			mapObj = obj.(map[string]interface{})
		} else {
			mapObj = make(map[string]interface{})
			mapData["devops"] = mapObj
		}

		for key, val := range dataMap {
			mapObj[key] = val
		}
	}
	resultData, _ := yaml.Marshal(mapData)
	return string(resultData)
}
