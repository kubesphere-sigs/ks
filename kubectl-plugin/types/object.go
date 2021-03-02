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
