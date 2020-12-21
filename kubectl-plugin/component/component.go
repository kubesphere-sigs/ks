package component

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	kstypes "github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"os"
	"os/signal"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

// NewComponentCmd returns a command to manage components of KubeSphere
func NewComponentCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "component",
		Aliases: []string{"com"},
		Short:   "Manage the components of KubeSphere",
	}

	cmd.AddCommand(NewComponentEnableCmd(client),
		NewComponentEditCmd(client),
		NewComponentResetCmd(client),
		NewComponentWatchCmd(client),
		NewComponentLogCmd(client))
	return
}

// Option is the common option for component command
type Option struct {
	Name    string
	Release bool
	Tag     string
	Client  dynamic.Interface
}

// ResetOption is the option for component reset command
type ResetOption struct {
	Option

	ResetAll bool
}

// WatchOption is the option for component watch command
type WatchOption struct {
	Option

	Watch       bool
	WatchImage  string
	WatchTag    string
	WatchDeploy string

	Registry         string
	RegistryUsername string
	RegistryPassword string
	PrivateRegistry  string
	PrivateAsLocal   bool
}

// NewComponentEnableCmd returns a command to enable (or disable) a component by name
func NewComponentEnableCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "enable",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("not supported yet")
			return nil
		},
	}
	return
}

// NewComponentWatchCmd returns a command to enable (or disable) a component by name
func NewComponentWatchCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &WatchOption{
		Option: Option{
			Client: client,
		},
	}
	cmd = &cobra.Command{
		Use:   "watch",
		Short: "Update images of ks-apiserver, ks-controller-manager, ks-console",
		RunE:  opt.watchRunE,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opt.Release, "release", "r", true,
		"Indicate if you want to update KubeSphere deploy image to release. Released images come from kubesphere/xxx. Otherwise images come from kubespheredev/xxx")
	flags.StringVarP(&opt.Tag, "tag", "t", kstypes.KsVersion,
		"The tag of KubeSphere deploys")
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

func (o *WatchOption) getDigest(image, tag string) string {
	dClient := kstypes.DockerClient{
		Image:           image,
		Registry:        o.Registry,
		PrivateRegistry: o.PrivateRegistry,
	}
	token := dClient.GetToken()
	dClient.Token = token
	return dClient.GetDigest(tag)
}

func (o *WatchOption) watchRunE(cmd *cobra.Command, args []string) (err error) {
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
				_, err = o.Client.Resource(kstypes.GetDeploySchema()).Namespace("kubesphere-system").Patch(ctx,
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
	return nil
}

func (o *WatchOption) getFullImagePath(image string) string {
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

// NewComponentResetCmd returns a command to enable (or disable) a component by name
func NewComponentResetCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &ResetOption{
		Option: Option{
			Client: client,
		},
	}
	cmd = &cobra.Command{
		Use:   "reset",
		Short: "reset the component by name",
		RunE:  opt.resetRunE,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opt.Release, "release", "r", true,
		"Indicate if you want to update KubeSphere deploy image to release. Released images come from KubeSphere/xxx. Otherwise images come from kubespheredev/xxx")
	flags.StringVarP(&opt.Tag, "tag", "t", kstypes.KsVersion,
		"The tag of KubeSphere deploys")
	flags.BoolVarP(&opt.ResetAll, "all", "a", false,
		"Indicate if you want to all supported components")
	flags.StringVarP(&opt.Name, "name", "n", "",
		"The name of target component which you want to reset. This does not work if you provide flag --all")
	return
}

func (o *ResetOption) resetRunE(cmd *cobra.Command, args []string) (err error) {
	if o.Tag == "" {
		// let users choose it if the tag option is empty
		dc := kstypes.DockerClient{
			Image: "kubesphere/ks-apiserver",
		}

		var tags *kstypes.DockerTags
		if tags, err = dc.GetTags(); err != nil {
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
	} else {
		o.Tag = "latest"
	}

	if o.ResetAll {
		o.Name = "apiserver"
		if err = o.updateBy(imageOrg, o.Tag); err != nil {
			return
		}

		o.Name = "controller"
		if err = o.updateBy(imageOrg, o.Tag); err != nil {
			return
		}

		o.Name = "console"
		if err = o.updateBy(imageOrg, o.Tag); err != nil {
			return
		}
	} else {
		err = o.updateBy(imageOrg, o.Tag)
	}
	return
}

func (o *Option) getNsAndName(component string) (ns, name string) {
	ns = "kubesphere-system"
	switch o.Name {
	case "apiserver":
		name = "ks-apiserver"
	case "controller", "controller-manager":
		name = "ks-controller-manager"
	case "console":
		name = "ks-console"
	}
	return
}

func (o *Option) getResourceType(component string) schema.GroupVersionResource {
	switch o.Name {
	default:
		fallthrough
	case "apiserver", "controller", "controller-manager", "console":
		return kstypes.GetDeploySchema()
	}
}

func (o *Option) updateBy(image, tag string) (err error) {
	ns, name := o.getNsAndName(o.Name)
	err = o.updateDeploy(ns, name, fmt.Sprintf("%s/%s", image, name), o.Tag)
	return
}

func (o *Option) updateDeploy(ns, name, image, tag string) (err error) {
	client := o.Client

	dClient := kstypes.DockerClient{
		Image: image,
	}
	token := dClient.GetToken()
	dClient.Token = token
	digest := dClient.GetDigest(tag)

	image = fmt.Sprintf("%s:%s@%s", image, tag, digest)
	fmt.Println("prepare to patch image", image)

	ctx := context.TODO()
	_, err = client.Resource(kstypes.GetDeploySchema()).Namespace(ns).Patch(ctx,
		name, types.JSONPatchType,
		[]byte(fmt.Sprintf(`[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "%s"}]`, image)),
		metav1.PatchOptions{})
	return
}

// NewComponentLogCmd returns a command to enable (or disable) a component by name
func NewComponentLogCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "log",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("not supported yet")
			return nil
		},
	}
	return
}

// NewComponentEditCmd returns a command to enable (or disable) a component by name
func NewComponentEditCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &Option{
		Client: client,
	}
	cmd = &cobra.Command{
		Use:     "edit",
		Short:   "edit the target component",
		PreRunE: opt.editPreRunE,
		RunE:    opt.editRunE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.Name, "name", "n", "",
		"The name of target component which you want to reset.")
	return
}

func (o *Option) editPreRunE(cmd *cobra.Command, args []string) (err error) {
	if len(args) > 0 {
		o.Name = args[0]
	}

	if o.Name == "" {
		err = fmt.Errorf("please provide the name of component")
	}
	return
}

func (o *Option) editRunE(cmd *cobra.Command, args []string) (err error) {
	var rawPip *unstructured.Unstructured
	var data []byte

	ctx := context.TODO()
	ns, name := o.getNsAndName(o.Name)
	resource := o.getResourceType(o.Name)

	buf := bytes.NewBuffer(data)
	if rawPip, err = o.Client.Resource(resource).Namespace(ns).Get(ctx, name, metav1.GetOptions{}); err == nil {
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

	if err = yaml.Unmarshal([]byte(content), rawPip); err == nil {
		_, err = o.Client.Resource(resource).Namespace(ns).Update(context.TODO(), rawPip, metav1.UpdateOptions{})
	}
	return
}
