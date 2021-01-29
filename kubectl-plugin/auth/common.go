package auth

import "gopkg.in/yaml.v3"

func updateAuthentication(yamlf, name, target string) string {
	targetobj := make(map[string]interface{}, 0)
	_ = yaml.Unmarshal([]byte(target), targetobj)
	return updateAuthWithObj(yamlf, name, targetobj)
}

func updateAuthWithObj(yamlf, name string, target map[string]interface{}) string {
	mapData := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(yamlf), mapData); err == nil {
		var obj interface{}
		var ok bool
		var mapObj map[string]interface{}
		if obj, ok = mapData["authentication"]; ok {
			mapObj = obj.(map[string]interface{})
		} else {
			return ""
		}

		if obj, ok = mapObj["oauthOptions"]; ok {
			mapObj = obj.(map[string]interface{})
		} else {
			oauthOptions := make(map[string]interface{}, 3)
			oauthOptions["accessTokenMaxAge"] = "1h"
			oauthOptions["accessTokenInactivityTimeout"] = "30m"
			mapObj["oauthOptions"] = oauthOptions
			mapObj = oauthOptions
		}

		targetArray := make([]interface{}, 0)
		found := false
		if obj, ok = mapObj["identityProviders"]; !ok {
			mapObj["identityProviders"] = &targetArray
		} else {
			array := obj.([]interface{})
			targetArray = array
			for i := range array {
				if array[i].(map[string]interface{})["name"] == name {
					found = true
					array[i] = target
					break
				}
			}
		}

		if !found {
			targetArray = append(targetArray, target)
			mapObj["identityProviders"] = targetArray
		}
	}
	resultData, _ := yaml.Marshal(mapData)
	return string(resultData)
}
