package types

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// GetObjectFromYaml returns the Unstructured object from a YAML
func GetObjectFromYaml(yamlText string) (obj *unstructured.Unstructured, err error) {
	obj = &unstructured.Unstructured{}
	err = yaml.Unmarshal([]byte(yamlText), obj)
	return
}

// GetObjectFromInterface returns the Unstructured object from a interface
func GetObjectFromInterface(raw interface{}) (obj *unstructured.Unstructured, err error) {
	var data []byte
	if data, err = yaml.Marshal(raw); err == nil {
		obj = &unstructured.Unstructured{}
		err = yaml.Unmarshal(data, obj)
	}
	return
}
