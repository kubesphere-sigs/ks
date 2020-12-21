package types

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// DockerClient is a simple Docker client
type DockerClient struct {
	Image           string
	Token           string
	Registry        string
	PrivateRegistry string
}

// DockerTags represents the docker tag list
type DockerTags struct {
	Name string
	Tags []string
}

// GetTags returns the tag list
func (d *DockerClient) GetTags() (tags *DockerTags, err error) {
	client := http.Client{}

	token := d.GetToken()
	d.Token = token

	var req *http.Request
	if req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("https://index.docker.io/v2/%s/tags/list", d.Image), nil); err != nil {
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.Token))
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	var rsp *http.Response
	if rsp, err = client.Do(req); err == nil && rsp != nil && rsp.StatusCode == http.StatusOK {
		var data []byte
		if data, err = ioutil.ReadAll(rsp.Body); err == nil {
			if err = json.Unmarshal(data, tags); err != nil {
				err = fmt.Errorf("unexpected docker image tag data, %#v", err)
			}
		}
	}
	return
}

// GetDigest returns the digest of the specific image tag
func (d *DockerClient) GetDigest(tag string) string {
	client := http.Client{}

	if tag == "" {
		tag = "latest"
	}

	var api string
	switch d.Registry {
	default:
		fallthrough
	case "docker":
		api = fmt.Sprintf("https://index.docker.io/v2/%s/manifests/%s", d.Image, tag)
	case "aliyun":
		api = fmt.Sprintf("https://registry.cn-beijing.aliyuncs.com/v2/%s/manifests/%s", d.Image, tag)
	case "qingcloud":
		api = fmt.Sprintf("https://dockerhub.qingcloud.com/v2/%s/manifests/%s", d.Image, tag)
	case "private":
		api = fmt.Sprintf("http://%s/v2/%s/manifests/%s", d.PrivateRegistry, d.Image, tag)
	}

	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodGet, api, nil); err != nil {
		return ""
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.Token))
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	if rsp, err := client.Do(req); err == nil && rsp != nil {
		if rsp.StatusCode != http.StatusOK {
			fmt.Println(req)
			if data, err := ioutil.ReadAll(rsp.Body); err == nil {
				fmt.Println(string(data))
			}
		}
		return rsp.Header.Get("Docker-Content-Digest")
	}
	return ""
}

type token struct {
	Token string
}

// GetToken returns the token of target docker image provider
func (d *DockerClient) GetToken() string {
	var api string
	switch d.Registry {
	default:
		fallthrough
	case "docker":
		api = fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", d.Image)
	case "aliyun":
		api = fmt.Sprintf("https://dockerauth.cn-beijing.aliyuncs.com/auth?&service=registry.aliyuncs.com:cn-beijing&scope=repository:%s:pull", d.Image)
	case "qingcloud":
		api = fmt.Sprintf("https://dockerauth.qingcloud.com:6000/auth?&service=dockerhub.qingcloud.com&scope=repository:%s:pull", d.Image)
	case "private":
		api = fmt.Sprintf("http://%s/auth?&service=%s&scope=repository:%s:pull", d.PrivateRegistry, d.PrivateRegistry, d.Image)
	}

	if req, err := http.NewRequest(http.MethodGet, api, nil); err == nil {
		httpClient := http.Client{}

		if rsp, err := httpClient.Do(req); err == nil && rsp.StatusCode == http.StatusOK {
			if data, err := ioutil.ReadAll(rsp.Body); err == nil {
				token := token{}
				if err = json.Unmarshal(data, &token); err == nil {
					return token.Token
				}
			}
		}
	}
	return ""
}
