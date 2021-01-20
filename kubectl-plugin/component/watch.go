package component

import (
	"context"
	"fmt"
	kstypes "github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"
)

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
		Example: `In order to make it be simple, please add the following environment variables
export KS_PRIVATE_LOCAL=192.168.0.8
export KS_REPO=139.198.3.176:32678
ks ks com watch --watch-deploy apiserver --watch-tag fix-pipe-list --registry private`,
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
		`a private registry, for example: docker run -d -p 5000:5000 --restart always --name registry registry:2
take value from environment 'KS_REPO' if you don't set it`)
	flags.StringVarP(&opt.PrivateLocal, "private-local", "", "127.0.0.1",
		`The local address of registry
take value from environment 'KS_PRIVATE_LOCAL' if you don't set it`)
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
