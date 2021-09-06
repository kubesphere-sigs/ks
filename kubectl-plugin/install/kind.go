package install

import (
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/linuxsuren/http-downloader/pkg/installer"
	"github.com/spf13/cobra"
	"html/template"
	"os"
	"path"
	"runtime"
	"sync"
)

func newInstallWithKindCmd() (cmd *cobra.Command) {
	opt := &kindOption{}
	cmd = &cobra.Command{
		Use:   "kind",
		Short: "Install KubeSphere with kind",
		Example: `ks install kind --components DevOps
ks install kind --nightly latest --components DevOps`,
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.name, "name", "n", "kind",
		"The name of kind")
	flags.StringVarP(&opt.version, "version", "v", "v1.18.2",
		"The version for Kubernetes")
	flags.StringToStringVarP(&opt.portMappings, "portMappings", "", map[string]string{"30880": "30881",
		"30180": "30181"}, "The extraPortMappings")
	flags.StringVarP(&opt.ksVersion, "ksVersion", "", types.KsVersion,
		"The version of KubeSphere")
	flags.StringSliceVarP(&opt.components, "components", "", []string{},
		"Which components will enable")
	flags.BoolVarP(&opt.Reset, "reset", "", false, "")
	flags.StringVarP(&opt.Nightly, "nightly", "", "",
		"Supported date format is '20200101', or you can use 'latest' which means yesterday")
	flags.BoolVarP(&opt.fetch, "fetch", "", true,
		"Indicate if fetch the latest config of tools")

	_ = cmd.RegisterFlagCompletionFunc("components", common.ArrayCompletion("DevOps"))
	return
}

func (o *kindOption) reset(cmd *cobra.Command, args []string) (err error) {
	var tag string
	if o.Nightly, tag = common.GetNightlyTag(o.Nightly); tag == "" {
		return
	}

	var wg sync.WaitGroup

	images := map[string]string{
		"kubesphere/ks-installer":          tag,
		"kubesphere/ks-apiserver":          tag,
		"kubesphere/ks-controller-manager": tag,
		"kubesphere/ks-console":            tag,
	}

	for image, version := range images {
		if err = pullAndLoadImageSync(fmt.Sprintf("%s:%s", image, version), o.name, &wg); err != nil {
			return
		}
	}

	wg.Wait()
	commander := Commander{}
	if err = commander.execCommand("kubectl", "ks", "com", "reset", "--nightly", o.Nightly, "-a"); err != nil {
		return
	}
	return
}

func (o *kindOption) preRunE(_ *cobra.Command, _ []string) (err error) {
	is := installer.Installer{
		Provider: "github",
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Fetch:    o.fetch,
	}
	err = is.CheckDepAndInstall(map[string]string{
		"kind":   "kind",
		"docker": "docker",
	})
	return
}

func (o *kindOption) runE(cmd *cobra.Command, args []string) (err error) {

	kindImage := fmt.Sprintf("kindest/node:%s", o.version)

	kindConfig := KindConfig{
		Image:        kindImage,
		PortMappings: o.portMappings,
	}

	kindConfigF := path.Join(os.TempDir(), "config.yaml")

	defer func() {
		err := os.Remove(kindConfigF)
		if err != nil {
			return
		}
	}()

	if err != writeConfigFile(kindConfigF, kindConfig) {
		return
	}

	commander := Commander{}
	if err = commander.execCommand("kind", "create", "cluster", "--config", kindConfigF, "--name", o.name); err != nil {
		return
	}

	if err = commander.execCommand("kubectl", "cluster-info", " --context", fmt.Sprintf("kind-%s", o.name)); err != nil {
		return
	}

	if err = o.loadCoreImageOfKS(); err != nil {
		return
	}

	if err = commander.execCommand("kubectl", "apply", "-f", fmt.Sprintf("https://github.com/kubesphere/ks-installer/releases/download/%s/kubesphere-installer.yaml", o.ksVersion)); err != nil {
		return
	}

	if err = commander.execCommand("kubectl", "apply", "-f", fmt.Sprintf("https://github.com/kubesphere/ks-installer/releases/download/%s/cluster-configuration.yaml", o.ksVersion)); err != nil {
		return
	}

	fmt.Println(`kubectl -n kubesphere-system patch deploy ks-installer --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"IfNotPresent"}]'`)

	for _, com := range o.components {
		switch com {
		case "DevOps":
			err = o.loadImagesOfDevOps()
		}
	}

	if o.Reset {
		err = o.reset(cmd, args)
	}
	return
}

func (o *kindOption) loadImagesOfDevOps() (err error) {
	var wg sync.WaitGroup

	images := map[string]string{
		"kubesphere/jenkins-uc":     "v3.0.0",
		"jenkins/jenkins":           "2.176.2",
		"jenkins/jnlp-slave":        "3.27-1",
		"kubesphere/builder-base":   "v2.1.0",
		"kubesphere/builder-go":     "v2.1.0",
		"kubesphere/builder-maven":  "v2.1.0",
		"kubesphere/builder-nodejs": "v2.1.0",
	}

	for image, tag := range images {
		if err = pullAndLoadImageSync(fmt.Sprintf("%s:%s", image, tag), o.name, &wg); err != nil {
			return
		}
	}

	wg.Wait()
	return
}

func (o *kindOption) loadCoreImageOfKS() (err error) {
	var wg sync.WaitGroup

	images := map[string]string{
		"kubesphere/ks-installer":          o.ksVersion,
		"kubesphere/ks-apiserver":          o.ksVersion,
		"kubesphere/ks-console":            o.ksVersion,
		"kubesphere/ks-controller-manager": o.ksVersion,
		"minio/minio":                      "RELEASE.2019-08-07T01-59-21Z",
		"mysql":                            "8.0.11",
		"osixia/openldap":                  "1.3.0",
	}

	for image, tag := range images {
		if err = pullAndLoadImageSync(fmt.Sprintf("%s:%s", image, tag), o.name, &wg); err != nil {
			return
		}
	}

	wg.Wait()
	return
}

func pullAndLoadImageSync(image string, kindName string, wg *sync.WaitGroup) (err error) {
	wg.Add(1)
	go func(imageName string, kindClusterName string, wgInner *sync.WaitGroup) {
		_ = pullAndLoadImage(imageName, kindClusterName)
		wgInner.Done()
	}(image, kindName, wg)
	return
}

func pullAndLoadImage(image string, kindName string) (err error) {
	commander := Commander{}
	if err = commander.execCommand("docker", "pull", image); err == nil {
		err = commander.execCommand("kind", "load", "docker-image", "--name", kindName, image)
	}
	return
}

func writeConfigFile(filename string, kindConfig KindConfig) (err error) {
	kindTemplate, err := template.New("config").Parse(`
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: {{ .Image }}
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
{{- range $k, $v := .PortMappings }}
  - containerPort: {{$k}}
    hostPort: {{$v}}
    protocol: TCP
{{- end }}
`)
	if err != nil {
		fmt.Println("failed to write kind config file", err)
		return
	}
	f, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0600)
	if err = kindTemplate.Execute(f, kindConfig); err != nil {
		fmt.Println("failed to render kind template", err)
		return
	}

	return
}

type kindOption struct {
	name         string
	version      string
	portMappings map[string]string
	ksVersion    string
	components   []string
	fetch        bool

	Reset   bool
	Nightly string
}

// KindConfig config template variables
type KindConfig struct {
	Image        string
	PortMappings map[string]string
}
