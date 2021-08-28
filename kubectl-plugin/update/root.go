package update

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	types2 "github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

type updateCmdOption struct {
	Release     bool
	Tag         string
	Watch       bool
	WatchImage  string
	WatchTag    string
	WatchDeploy string

	Registry         string
	RegistryUsername string
	RegistryPassword string
	PrivateRegistry  string
	PrivateAsLocal   bool
	Client           dynamic.Interface
}

// NewUpdateCmd returns a command of update
func NewUpdateCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := updateCmdOption{
		Client: client,
	}

	cmd = &cobra.Command{
		Use:        "update",
		Short:      "Update images of ks-apiserver, ks-controller-manager, ks-console",
		Aliases:    []string{"up"},
		Deprecated: "This command will be removed after v0.1.0. Please use kubectl ks component xxx instead.",
		PreRun:     opt.preRun,
		Args:       opt.args,
		RunE:       opt.RunE,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opt.Release, "release", "r", true,
		"Indicate if you want to update Kubesphere deploy image to release. Released images come from kubesphere/xxx. Otherwise images come from kubespheredev/xxx")
	flags.StringVarP(&opt.Tag, "tag", "t", types2.KsVersion,
		"The tag of Kubesphere deploys")
	flags.BoolVarP(&opt.Watch, "watch", "w", false,
		"Watch a container image then update it")
	flags.StringVarP(&opt.WatchDeploy, "watch-deploy", "", "",
		"Watch a deploy then update it")
	flags.StringVarP(&opt.WatchImage, "watch-image", "", "",
		"which image you want to watch")
	flags.StringVarP(&opt.WatchTag, "watch-tag", "", "",
		"which image tag you want to watch")
	flags.StringVarP(&opt.Registry, "registry", "", "docker",
		"supported list [docker, aliyun, qingcloud, private], we only support beijing area of aliyun")
	flags.StringVarP(&opt.PrivateRegistry, "private-registry", "", "",
		"a private registry, for example: docker run -d -p 5000:5000 --restart always --name registry registry:2 ")
	flags.BoolVarP(&opt.PrivateAsLocal, "private-as-local", "", true,
		"use 127.0.0.1 as the private registry host")
	return
}

func (o *updateCmdOption) args(cmd *cobra.Command, args []string) (err error) {
	if o.Watch {
		if o.WatchDeploy == "" || o.WatchImage == "" || o.WatchTag == "" {
			err = fmt.Errorf("--watch-deploy, --watch-image, --image-tag cannot be empty")
			return
		}

		if o.Registry == "private" && o.PrivateRegistry == "" {
			err = fmt.Errorf("--private-regitry cannot be empty if you want watch a private registry")
		}
	}
	return
}

func (o *updateCmdOption) preRun(cmd *cobra.Command, args []string) {
	if o.Release {
		o.Tag = types2.KsVersion
	} else {
		o.Tag = "latest"
	}
}

func (o *updateCmdOption) getDigest(image, tag string) string {
	dClient := DockerClient{
		Image:           image,
		Registry:        o.Registry,
		PrivateRegistry: o.PrivateRegistry,
	}
	token := dClient.getToken()
	dClient.Token = token
	return dClient.getDigest(tag)
}

func (o *updateCmdOption) getFullImagePath(image string) string {
	switch o.Registry {
	default:
		fallthrough
	case "docker":
		return image
	case "aliyun":
		return fmt.Sprintf("registry.cn-beijing.aliyuncs.com/%s", image)
	case "qingcloud":
		return fmt.Sprintf("dockerhub.qingcloud.com/%s", image)
	case "private":
		if o.PrivateAsLocal {
			regAndPort := strings.Split(o.PrivateRegistry, ":")
			return fmt.Sprintf("127.0.0.1:%s/%s", regAndPort[1], image)
		}
		return fmt.Sprintf("%s/%s", o.PrivateRegistry, image)
	}
}

func (o *updateCmdOption) RunE(cmd *cobra.Command, args []string) (err error) {
	if o.Watch {
		var currentDigest string
		digestChain := make(chan string)
		go func(digestChain chan<- string) {
			for {
				digestChain <- o.getDigest(o.WatchImage, o.WatchTag)
				time.Sleep(time.Second * 2)
			}
		}(digestChain)

		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, os.Kill)
		signal.Notify(sigChan, os.Interrupt)

		for {
			select {
			case digest := <-digestChain:
				if digest != currentDigest && digest != "" {
					fmt.Println("prepare to patch image, new digest is", digest, "old digest is", currentDigest)
					fmt.Println("image", o.getFullImagePath(fmt.Sprintf("%s:%s@%s", o.WatchImage, o.WatchTag, digest)))
					currentDigest = digest

					ctx := context.TODO()
					_, err = o.Client.Resource(types2.GetDeploySchema()).Namespace("kubesphere-system").Patch(ctx,
						o.WatchDeploy, types.JSONPatchType,
						[]byte(fmt.Sprintf(`[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "%s"}]`,
							o.getFullImagePath(fmt.Sprintf("%s:%s@%s", o.WatchImage, o.WatchTag, digest)))),
						metav1.PatchOptions{})
				}
			case sig := <-sigChan:
				fmt.Println(sig)
				return
			}
		}
	}

	if o.Tag == "" {
		dc := DockerClient{
			Image: "kubesphere/ks-apiserver",
		}

		var tags *DockerTags
		if tags, err = dc.getTags(); err != nil {
			err = fmt.Errorf("cannot get the tags, %#v", err)
			return
		}

		prompt := &survey.Select{
			Message: "Please select the tag which you want to check:",
			Options: tags.Tags,
		}
		if err = survey.AskOne(prompt, &o.Tag); err != nil {
			return
		}
	}

	imageOrg := "kubespheredev"
	if o.Release {
		imageOrg = "kubesphere"
	}

	_ = o.updateDeploy("kubesphere-system", "ks-apiserver", fmt.Sprintf("%s/ks-apiserver", imageOrg), o.Tag)
	_ = o.updateDeploy("kubesphere-system", "ks-controller-manager", fmt.Sprintf("%s/ks-controller-manager", imageOrg), o.Tag)
	_ = o.updateDeploy("kubesphere-system", "ks-console", fmt.Sprintf("%s/ks-console", imageOrg), o.Tag)
	return
}

func (o *updateCmdOption) updateDeploy(ns, name, image, tag string) (err error) {
	client := o.Client

	dClient := DockerClient{
		Image: image,
	}
	token := dClient.getToken()
	dClient.Token = token
	digest := dClient.getDigest(tag)

	image = fmt.Sprintf("%s:%s@%s", image, tag, digest)
	fmt.Println("prepare to patch image", image)

	ctx := context.TODO()
	_, err = client.Resource(types2.GetDeploySchema()).Namespace(ns).Patch(ctx,
		name, types.JSONPatchType,
		[]byte(fmt.Sprintf(`[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "%s"}]`, image)),
		metav1.PatchOptions{})
	return
}

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

func (d *DockerClient) getTags() (tags *DockerTags, err error) {
	client := http.Client{}

	token := d.getToken()
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

func (d *DockerClient) getDigest(tag string) string {
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

func (d *DockerClient) getToken() string {
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
