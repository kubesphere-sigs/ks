package component

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/linuxsuren/ks/kubectl-plugin/common"
	kstypes "github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

// NewComponentCmd returns a command to manage components of KubeSphere
func NewComponentCmd(client dynamic.Interface, clientset *kubernetes.Clientset) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "component",
		Aliases: []string{"com"},
		Short:   "Manage the components of KubeSphere",
	}

	cmd.AddCommand(NewComponentEnableCmd(client),
		NewComponentEditCmd(client),
		NewComponentResetCmd(client),
		NewComponentWatchCmd(client),
		NewComponentLogCmd(client, clientset))
	return
}

// Option is the common option for component command
type Option struct {
	Name      string
	Release   bool
	Tag       string
	Client    dynamic.Interface
	Clientset *kubernetes.Clientset
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
	PrivateLocal     string
}

// EnableOption is the option for component enable command
type EnableOption struct {
	Option

	Edit   bool
	Toggle bool
}

// NewComponentEnableCmd returns a command to enable (or disable) a component by name
func NewComponentEnableCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &EnableOption{
		Option: Option{
			Client: client,
		},
	}
	cmd = &cobra.Command{
		Use:     "enable",
		Short:   "Enable or disable the specific KubeSphere component",
		PreRunE: opt.enablePreRunE,
		RunE:    opt.enableRunE,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opt.Edit, "edit", "e", false,
		"Indicate if you want to edit it instead of enable/disable a specified one. This flag will make others not work.")
	flags.BoolVarP(&opt.Toggle, "toggle", "t", false,
		"Indicate if you want to disable a component")
	flags.StringVarP(&opt.Name, "name", "n", "",
		"The name of target component which you want to enable/disable.")
	return
}

func (o *EnableOption) enablePreRunE(cmd *cobra.Command, args []string) (err error) {
	if o.Edit {
		return
	}

	return o.componentNameCheck(cmd, args)
}

func (o *EnableOption) enableRunE(cmd *cobra.Command, args []string) (err error) {
	if o.Edit {
		err = common.UpdateWithEditor(kstypes.GetClusterConfiguration(), "kubesphere-system", "ks-installer", o.Client)
	} else {
		enabled := strconv.FormatBool(!o.Toggle)
		ns, name := "kubesphere-system", "ks-installer"
		var patchTarget string
		switch o.Name {
		case "devops", "alerting", "auditing", "events", "logging", "metrics_server", "networkpolicy", "notification", "openpitrix", "servicemesh":
			patchTarget = o.Name
		default:
			err = fmt.Errorf("not support [%s] yet", o.Name)
			return
		}

		patch := fmt.Sprintf(`[{"op": "replace", "path": "/spec/%s/enabled", "value": %s}]`, patchTarget, enabled)
		ctx := context.TODO()
		_, err = o.Client.Resource(kstypes.GetClusterConfiguration()).Namespace(ns).Patch(ctx,
			name, types.JSONPatchType,
			[]byte(patch),
			metav1.PatchOptions{})
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
		Use:     "watch",
		Short:   "Update images of ks-apiserver, ks-controller-manager, ks-console",
		PreRunE: opt.watchPreRunE,
		RunE:    opt.watchRunE,
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
	flags.StringVarP(&opt.PrivateLocal, "private-local", "", "127.0.0.1",
		"The local address of registry")
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

func (o *WatchOption) watchPreRunE(cmd *cobra.Command, args []string) (err error) {
	if o.PrivateRegistry == "" {
		o.PrivateRegistry = os.Getenv("kS_PRIVATE_REG")
	}

	switch o.WatchDeploy {
	case "api", "apiserver":
		o.WatchDeploy = "ks-apiserver"
		if o.WatchImage == "" {
			o.WatchImage = "kubespheredev/ks-apiserver"
		}
	case "ctl", "ctrl", "controller":
		o.WatchDeploy = "ks-controller"
		if o.WatchImage == "" {
			o.WatchImage = "ks-controller-manager"
		}
	case "console":
		o.WatchDeploy = "ks-console"
		if o.WatchImage == "" {
			o.WatchImage = "kubespheredev/ks-console"
		}
	}

	if o.WatchTag == "" {
		if data, err := exec.Command("git", "branch", "--show-current").Output(); err == nil {
			tag := strings.TrimSpace(string(data))
			if tag == "master" {
				tag = "latest"
			}
			o.WatchTag = strings.ReplaceAll(tag, "/", "-")
		}
	}

	if o.PrivateRegistry == "" {
		o.PrivateRegistry = os.Getenv("KS_REPO")
	}

	if local, ok := os.LookupEnv("KS_PRIVATE_LOCAL"); ok && o.PrivateLocal == "127.0.0.1" {
		o.PrivateLocal = local
	}

	// check the necessary options
	if o.Registry == "private" && o.PrivateRegistry == "" {
		err = fmt.Errorf("--private-registry cannot be empty if you have --registry=private")
		return
	}
	if o.WatchImage == "" {
		err = fmt.Errorf("--watch-image cannot be empty")
		return
	}
	return
}

func (o *WatchOption) watchRunE(cmd *cobra.Command, args []string) (err error) {
	cmd.Println("start to watch", o.getFullImagePath(fmt.Sprintf("%s:%s", o.WatchImage, o.WatchTag)))

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
				if _, err = o.Client.Resource(kstypes.GetDeploySchema()).Namespace("kubesphere-system").Patch(ctx,
					o.WatchDeploy, types.JSONPatchType,
					[]byte(fmt.Sprintf(`[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "%s"}]`,
						o.getFullImagePath(fmt.Sprintf("%s:%s@%s", o.WatchImage, o.WatchTag, digest)))),
					metav1.PatchOptions{}); err != nil {
					cmd.PrintErrln(err)
				}
			}
		case <-sigChan:
			return
		}
	}
	return
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
		regAndPort := strings.Split(o.PrivateRegistry, ":")
		return fmt.Sprintf("%s:%s/%s", o.PrivateLocal, regAndPort[1], image)
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
	case "installer":
		name = "ks-installer"
	case "jenkins":
		name = "ks-jenkins"
		ns = "kubesphere-devops-system"
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

// LogOption is the option for component log command
type LogOption struct {
	Option

	Follow bool
	Tail   int64
}

// NewComponentLogCmd returns a command to enable (or disable) a component by name
func NewComponentLogCmd(client dynamic.Interface, clientset *kubernetes.Clientset) (cmd *cobra.Command) {
	opt := &LogOption{
		Option: Option{
			Clientset: clientset,
			Client:    client,
		},
	}
	cmd = &cobra.Command{
		Use:     "log",
		Short:   "Output the log of KubeSphere component",
		PreRunE: opt.componentNameCheck,
		RunE:    opt.logRunE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.Name, "name", "n", "",
		"The name of target component which you want to reset.")
	flags.BoolVarP(&opt.Follow, "follow", "f", true,
		"Specify if the logs should be streamed.")
	flags.Int64VarP(&opt.Tail, "tail", "", 50,
		`Lines of recent log file to display.`)
	return
}

func (o *LogOption) logRunE(cmd *cobra.Command, args []string) (err error) {
	if o.Clientset == nil {
		err = fmt.Errorf("kubernetes clientset is nil")
		return
	}

	ctx := context.TODO()
	var ns, name string
	if ns, name = o.getNsAndName(o.Name); name == "" {
		err = fmt.Errorf("not supported yet: %s", o.Name)
		return
	}

	var data []byte
	buf := bytes.NewBuffer(data)
	var rawPip *unstructured.Unstructured
	deploy := &simpleDeploy{}
	if rawPip, err = o.Client.Resource(kstypes.GetDeploySchema()).Namespace(ns).Get(ctx, name, metav1.GetOptions{}); err == nil {
		enc := json.NewEncoder(buf)
		enc.SetIndent("", "    ")
		if err = enc.Encode(rawPip); err != nil {
			return
		}

		cmd.Println(buf)
		if err = json.Unmarshal(buf.Bytes(), deploy); err != nil {
			return
		}
	}

	var podList *v1.PodList
	var podName string
	if podList, err = o.Clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(deploy.Spec.Selector.MatchLabels).String(),
	}); err == nil {
		if len(podList.Items) > 0 {
			podName = podList.Items[0].Name
		}
	} else {
		return
	}

	if podName == "" {
		err = fmt.Errorf("cannot found the pod with deployment '%s'", name)
		return
	}

	if len(deploy.Spec.Selector.MatchLabels) > 0 {
		req := o.Clientset.CoreV1().Pods(ns).GetLogs(podName, &v1.PodLogOptions{
			Follow:    o.Follow,
			TailLines: &o.Tail,
		})
		var podLogs io.ReadCloser
		if podLogs, err = req.Stream(context.TODO()); err == nil {
			defer func() {
				_ = podLogs.Close()
			}()

			_, err = io.Copy(cmd.OutOrStdout(), podLogs)
		}
	}
	return
}

type simpleDeploy struct {
	Spec struct {
		Selector struct {
			MatchLabels map[string]string `json:"matchLabels"`
		} `json:"selector"`
	} `json:"spec"`
}

// NewComponentEditCmd returns a command to enable (or disable) a component by name
func NewComponentEditCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &Option{
		Client: client,
	}
	cmd = &cobra.Command{
		Use:     "edit",
		Short:   "edit the target component",
		PreRunE: opt.componentNameCheck,
		RunE:    opt.editRunE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.Name, "name", "n", "",
		"The name of target component which you want to reset.")
	return
}

func (o *Option) componentNameCheck(cmd *cobra.Command, args []string) (err error) {
	if len(args) > 0 {
		o.Name = args[0]
	}

	if o.Name == "" {
		err = fmt.Errorf("please provide the name of component")
	}
	return
}

func (o *Option) editRunE(cmd *cobra.Command, args []string) (err error) {
	ns, name := o.getNsAndName(o.Name)
	resource := o.getResourceType(o.Name)

	err = common.UpdateWithEditor(resource, ns, name, o.Client)
	return
}
