package component

import (
	"context"
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	kstypes "github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// NewComponentCmd returns a command to manage components of KubeSphere
func NewComponentCmd(client dynamic.Interface, clientset *kubernetes.Clientset) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "component",
		Aliases: []string{"com"},
		Short:   "Manage the components of KubeSphere",
	}

	cmd.AddCommand(newComponentEnableCmd(),
		NewComponentEditCmd(),
		NewComponentResetCmd(),
		NewComponentWatchCmd(),
		newComponentLogCmd(),
		newComponentsExecCmd(),
		newComponentsKillCmd(),
		newScaleCmd(),
		newComponentDescribeCmd())
	return
}

// Option is the common option for component command
type Option struct {
	Name    string
	Release bool
	Tag     string

	SonarQube      string
	SonarQubeToken string

	Client    dynamic.Interface
	Clientset *kubernetes.Clientset
}

// ResetOption is the option for component reset command
type ResetOption struct {
	Option

	ResetAll bool
	Nightly  string
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

func (o *Option) updateBy(image string) (err error) {
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
	var digest kstypes.ImageDigest
	if digest, err = dClient.GetDigestObj(tag); err != nil {
		return
	}

	if digest.Digest == "" {
		err = fmt.Errorf("cannot get the digest of image '%s:%s'", image, tag)
		return
	}

	image = fmt.Sprintf("%s:%s@%s", image, tag, digest.Digest)
	fmt.Printf("prepare to patch image: '%s'\nbuild data: %s\n", image, digest.Date)

	ctx := context.TODO()
	_, err = client.Resource(kstypes.GetDeploySchema()).Namespace(ns).Patch(ctx,
		name, types.JSONPatchType,
		[]byte(fmt.Sprintf(`[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "%s"}]`, image)),
		metav1.PatchOptions{})
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
func NewComponentEditCmd() (cmd *cobra.Command) {
	opt := &Option{}
	cmd = &cobra.Command{
		Use:     "edit",
		Short:   "Edit the target component",
		PreRunE: opt.componentNameCheck,
		RunE:    opt.editRunE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.Name, "name", "n", "",
		"The name of target component which you want to reset.")
	return
}

func (o *Option) componentNameCheck(cmd *cobra.Command, args []string) (err error) {
	ctx := cmd.Root().Context()
	o.Client = common.GetDynamicClient(ctx)
	o.Clientset = common.GetClientset(ctx)

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
