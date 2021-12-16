package component

import (
	"context"
	"fmt"
	kstypes "github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"strings"
)

const cfgFile = "local_config.yaml"

func integrateSonarQube(client dynamic.Interface, ns, name, sonarqubeURL, sonarqubeToken string) (err error) {
	ctx := context.TODO()
	var rawConfigMap *unstructured.Unstructured
	if rawConfigMap, err = client.Resource(kstypes.GetConfigMapSchema()).Namespace(ns).
		Get(ctx, name, metav1.GetOptions{}); err == nil {
		data := rawConfigMap.Object["data"]
		dataMap := data.(map[string]interface{})
		result := updateConsoleConfig(dataMap[cfgFile].(string), sonarqubeURL)
		if strings.TrimSpace(result) == "" {
			err = fmt.Errorf("error happend when parse kubesphere-config")
			return
		}

		ctx := context.TODO()
		dataMap[cfgFile] = result
		rawConfigMap.Object["data"] = dataMap
		if _, err = client.Resource(kstypes.GetConfigMapSchema()).Namespace(ns).Update(ctx,
			rawConfigMap, metav1.UpdateOptions{}); err != nil {
			return
		}

		patch := fmt.Sprintf(`[{"op": "add", "path": "/spec/devops/sonarqube", "value": {"externalSonarUrl":"%s","externalSonarToken":"%s"}}]`,
			sonarqubeURL, sonarqubeToken)
		fmt.Println(patch)
		if _, err = client.Resource(kstypes.GetClusterConfiguration()).Namespace(ns).Patch(ctx,
			"ks-installer", types.JSONPatchType,
			[]byte(patch),
			metav1.PatchOptions{}); err != nil {
			return
		}
	}
	return
}

func updateConsoleConfig(yamlf, sonarqubeURL string) string {
	mapData := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(yamlf), mapData); err == nil {
		var obj interface{}
		var ok bool
		var mapObj map[string]interface{}
		if obj, ok = mapData["client"]; ok {
			mapObj = obj.(map[string]interface{})
		} else {
			return ""
		}

		if obj, ok = mapObj["devops"]; ok {
			mapObj = obj.(map[string]interface{})
		} else {
			devopsOpt := make(map[string]interface{}, 1)
			mapObj["devops"] = devopsOpt
			mapObj = devopsOpt
		}

		mapObj["sonarqubeURL"] = sonarqubeURL
	}
	resultData, _ := yaml.Marshal(mapData)
	return string(resultData)
}
