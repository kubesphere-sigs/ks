package component

import (
	"context"
	"fmt"
	"github.com/linuxsuren/ks/kubectl-plugin/common"
	kstypes "github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"strconv"
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
		"The name of target component which you want to enable/disable. Please provide option --sonarqube if you want to enable SonarQube.")
	flags.StringVarP(&opt.SonarQube, "sonarqube", "", "",
		"The SonarQube URL")
	flags.StringVarP(&opt.SonarQube, "sonar", "", "",
		"The SonarQube URL")
	flags.StringVarP(&opt.SonarQubeToken, "sonarqube-token", "", "",
		"The token of SonarQube")

	// these are aliased options
	_ = flags.MarkHidden("sonar")
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
		case "sonarqube", "sonar":
			if o.SonarQube == "" || o.SonarQubeToken == "" {
				err = fmt.Errorf("SonarQube or token is empty, please provide --sonarqube")
			} else {
				name = "ks-console-config"
				err = integrateSonarQube(o.Client, ns, name, o.SonarQube, o.SonarQubeToken)
			}
			return
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
	var digest kstypes.ImageDigest
	if digest, err = dClient.GetDigestObj(tag); err != nil {
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
