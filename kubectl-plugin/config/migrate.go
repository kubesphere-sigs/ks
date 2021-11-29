package config

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
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

	patchData := make(map[string]interface{})

	kubesphereConfig, _, err := o.getKubeSphereConfig("kubesphere-config", "kubesphere-system")
	if err != nil {
		return err
	}
	if sonarQube, found, err := unstructured.NestedMap(kubesphereConfig, "sonarQube"); err != nil {
		return err
	} else if found {
		patchData["sonarQube"] = sonarQube
	}

	var password string
	if password, err = o.getDevOpsPassword(); password == "" {
		if err == nil {
			err = fmt.Errorf("the password of Jenkins is empty")
		} else {
			err = fmt.Errorf("the password of Jenkins is empty, it might caused by: %v", err)
		}
	} else if err == nil {
		patchData["devops"] = map[string]interface{}{
			"password": password,
		}
	}

	return o.updateKubeSphereConfig("devops-config", o.namespace, patchData)
}

func (o *migrateOption) updateKubeSphereConfig(name, namespace string, ksdataMap map[string]interface{}) error {
	kubeSphereConfig, rawConfigMap, err := o.getKubeSphereConfig(name, namespace)
	if err != nil {
		return fmt.Errorf("cannot found ConfigMap %s/%s, %v", namespace, name, err)
	}

	mergeMap(kubeSphereConfig, ksdataMap)
	kubeSphereConfigBytes, err := yaml.Marshal(kubeSphereConfig)
	if err != nil {
		return fmt.Errorf("cannot marshal KubeSphere configuration, %v", err)
	}

	rawConfigMap.Object["data"] = map[string]interface{}{
		"kubesphere.yaml": string(kubeSphereConfigBytes),
	}
	if _, err = o.client.Resource(types.GetConfigMapSchema()).
		Namespace(namespace).
		Update(context.TODO(), rawConfigMap, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
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

func (o *migrateOption) getKubeSphereConfig(configMapName, namespace string) (map[string]interface{}, *unstructured.Unstructured, error) {
	kubesphereConfigCM, err := o.client.Resource((types.GetConfigMapSchema())).
		Namespace(namespace).
		Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("cannot found ConfigMap %s/%s, %v", namespace, configMapName, err)
	}
	kubesphereConfigYAMLString, found, err := unstructured.NestedString(kubesphereConfigCM.UnstructuredContent(), "data", "kubesphere.yaml")
	if err != nil {
		return nil, nil, err
	}
	if !found {
		return nil, nil, fmt.Errorf("cannot found 'kubesphere.yaml' configuration in ConfigMap %s/%s", namespace, configMapName)
	}
	kubesphereConfig := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(kubesphereConfigYAMLString), kubesphereConfig); err != nil {
		return nil, nil, err
	}
	return kubesphereConfig, kubesphereConfigCM, nil
}

// mergeMap merges patch map into main map.
// This function has a constriant that map types must be `map[string]interface{}`, including types of map value or map value's value.
func mergeMap(main map[string]interface{}, patch map[string]interface{}) {
	for patchKey, patchValue := range patch {
		if value, ok := main[patchKey]; ok {
			patchValueMap, patchValueOk := patchValue.(map[string]interface{})
			valueMap, valueOk := value.(map[string]interface{})
			if patchValueOk && valueOk {
				// recursive check
				mergeMap(valueMap, patchValueMap)
			} else {
				if reflect.TypeOf(value) == reflect.TypeOf(patchValue) {
					// set patch value directly if one of the value type are not map[string]interface{} and both types are equal
					main[patchKey] = patchValue
				}
			}
		} else {
			// set patch value directly if key in main is not found
			main[patchKey] = patchValue
		}
	}
}
