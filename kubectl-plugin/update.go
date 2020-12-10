package main

import (
	"context"
	"github.com/AlecAivazis/survey/v2"
	"k8s.io/apimachinery/pkg/types"

	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"net/http"
)

type updateCmdOption struct {
	Release bool
	Tag     string

	Client dynamic.Interface
}

// NewUpdateCmd returns a command of update
func NewUpdateCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := updateCmdOption{
		Client: client,
	}

	cmd = &cobra.Command{
		Use:     "update",
		Short:   "Update images of ks-apiserver, ks-controller-manager, ks-console",
		Aliases: []string{"up"},
		PreRun:  opt.PreRun,
		RunE:    opt.RunE,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opt.Release, "release", "r", true,
		"Indicate if you want to update Kubesphere deploy image to release. Released images come from kubesphere/xxx. Otherwise images come from kubespheredev/xxx")
	flags.StringVarP(&opt.Tag, "tag", "t", KS_VERSION,
		"The tag of Kubesphere deploys")
	return
}

func (o *updateCmdOption) PreRun(cmd *cobra.Command, args []string) {
	if o.Release {
		o.Tag = KS_VERSION
	} else {
		o.Tag = "latest"
	}
}

func (o *updateCmdOption) RunE(cmd *cobra.Command, args []string) (err error) {
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

	o.updateDeploy("kubesphere-system", "ks-apiserver", fmt.Sprintf("%s/ks-apiserver", imageOrg), o.Tag)
	o.updateDeploy("kubesphere-system", "ks-controller-manager", fmt.Sprintf("%s/ks-controller-manager", imageOrg), o.Tag)
	o.updateDeploy("kubesphere-system", "ks-console", fmt.Sprintf("%s/ks-console", imageOrg), o.Tag)
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
	_, err = client.Resource(GetDeploySchema()).Namespace(ns).Patch(ctx,
		name, types.JSONPatchType,
		[]byte(fmt.Sprintf(`[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "%s"}]`, image)),
		metav1.PatchOptions{})
	return
}

// DockerClient is a simple Docker client
type DockerClient struct {
	Image string
	Token string
}

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
	} else {
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
	}
	return
}

func (d *DockerClient) getDigest(tag string) string {
	client := http.Client{}

	if tag == "" {
		tag = "latest"
	}
	if req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://index.docker.io/v2/%s/manifests/%s", d.Image, tag), nil); err != nil {
		return ""
	} else {
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
	}
	return ""
}

type token struct {
	Token string
}

func (d *DockerClient) getToken() string {
	if rsp, err := http.Get(fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", d.Image)); err == nil && rsp.StatusCode == http.StatusOK {
		if data, err := ioutil.ReadAll(rsp.Body); err == nil {
			token := token{}
			if err = json.Unmarshal(data, &token); err == nil {
				return token.Token
			}
		}
	}

	return ""
}
