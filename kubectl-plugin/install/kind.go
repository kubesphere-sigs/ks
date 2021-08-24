package install

import (
	"fmt"
	"github.com/linuxsuren/http-downloader/pkg/installer"
	"github.com/linuxsuren/ks/kubectl-plugin/common"
	"github.com/spf13/cobra"
	"html/template"
	"os"
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
	flags.StringVarP(&opt.ksVersion, "ksVersion", "", "v3.0.0",
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
	if err = pullAndLoadImageSync(fmt.Sprintf("kubespheredev/ks-installer:%s", tag), &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync(fmt.Sprintf("kubespheredev/ks-apiserver:%s", tag), &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync(fmt.Sprintf("kubespheredev/ks-controller-manager:%s", tag), &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync(fmt.Sprintf("kubespheredev/ks-console:%s", tag), &wg); err != nil {
		return
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
	writeConfigFile("config.yaml", o.portMappings)
	commander := Commander{}
	if err = commander.execCommand("kind", "create", "cluster", "--image", "kindest/node:v1.18.2", "--config", "config.yaml", "--name", o.name); err != nil {
		return
	}

	if err = commander.execCommand("kubectl", "cluster-info", " --context", fmt.Sprintf("kind-%s", o.name)); err != nil {
		return
	}

	if err = loadCoreImageOfKS(); err != nil {
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
			err = loadImagesOfDevOps()
		}
	}

	if o.Reset {
		err = o.reset(cmd, args)
	}
	return
}

func loadImagesOfDevOps() (err error) {
	var wg sync.WaitGroup
	if err = pullAndLoadImageSync("kubesphere/jenkins-uc:v3.0.0", &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync("jenkins/jenkins:2.176.2", &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync("jenkins/jnlp-slave:3.27-1", &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync("kubesphere/builder-base:v2.1.0", &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync("kubesphere/builder-nodejs:v2.1.0", &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync("kubesphere/builder-go:v2.1.0", &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync("kubesphere/builder-maven:v2.1.0", &wg); err != nil {
		return
	}
	wg.Wait()
	return
}

func loadCoreImageOfKS() (err error) {
	var wg sync.WaitGroup
	if err = pullAndLoadImageSync("kubesphere/ks-installer:v3.0.0", &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync("kubesphere/ks-apiserver:v3.0.0", &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync("kubesphere/ks-controller-manager:v3.0.0", &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync("kubesphere/ks-console:v3.0.0", &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync("redis:5.0.5-alpine", &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync("osixia/openldap:1.3.0", &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync("minio/minio:RELEASE.2019-08-07T01-59-21Z", &wg); err != nil {
		return
	}
	if err = pullAndLoadImageSync("mysql:8.0.11", &wg); err != nil {
		return
	}
	wg.Wait()
	return
}

func pullAndLoadImageSync(image string, wg *sync.WaitGroup) (err error) {
	wg.Add(1)
	go func(imageName string, wgInner *sync.WaitGroup) {
		_ = pullAndLoadImage(imageName)
		wgInner.Done()
	}(image, wg)
	return
}

func pullAndLoadImage(image string) (err error) {
	commander := Commander{}
	if err = commander.execCommand("docker", "pull", image); err == nil {
		err = commander.execCommand("kind", "load", "docker-image", image)
	}
	return
}

func writeConfigFile(filename string, portMapping map[string]string) {
	kindTemplate, err := template.New("config").Parse(`
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
{{- range $k, $v := . }}
  - containerPort: {{$k}}
    hostPort: {{$v}}
    protocol: TCP
{{- end }}
`)
	if err != nil {
		fmt.Println("failed to write kind config file", err)
	}
	f, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0600)
	if err := kindTemplate.Execute(f, portMapping); err != nil {
		fmt.Println("failed to render kind template", err)
	}
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
