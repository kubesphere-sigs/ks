package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"os"
	"sigs.k8s.io/yaml"
)

// UpdateWithEditor update the k8s resources with a editor
func UpdateWithEditor(resource schema.GroupVersionResource, ns, name string, client dynamic.Interface) (err error) {
	var rawPip *unstructured.Unstructured
	var data []byte
	ctx := context.TODO()

	buf := bytes.NewBuffer(data)
	if rawPip, err = client.Resource(resource).Namespace(ns).Get(ctx, name, metav1.GetOptions{}); err == nil {
		enc := json.NewEncoder(buf)
		enc.SetIndent("", "    ")
		if err = enc.Encode(rawPip); err != nil {
			return
		}
	} else {
		err = fmt.Errorf("cannot get component, error: %#v", err)
		return
	}

	var yamlData []byte
	if yamlData, err = yaml.JSONToYAML(buf.Bytes()); err != nil {
		return
	}

	var fileName = "*.yaml"
	var content = string(yamlData)

	prompt := &survey.Editor{
		Message:       fmt.Sprintf("Edit component %s/%s", ns, name),
		FileName:      fileName,
		Default:       string(yamlData),
		HideDefault:   true,
		AppendDefault: true,
	}

	err = survey.AskOne(prompt, &content, survey.WithStdio(os.Stdin, os.Stdout, os.Stderr))
	if content == string(yamlData) {
		fmt.Println("Edit cancelled, no changes made.")
		return
	}

	if err = yaml.Unmarshal([]byte(content), rawPip); err == nil {
		_, err = client.Resource(resource).Namespace(ns).Update(context.TODO(), rawPip, metav1.UpdateOptions{})
	}
	return
}
