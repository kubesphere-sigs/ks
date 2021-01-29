package install

import (
	"fmt"
	"github.com/spf13/cobra"
	"html/template"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

func newInstallWithKindCmd() (cmd *cobra.Command) {
	opt := &kindOption{}
	cmd = &cobra.Command{
		Use:   "kind",
		Short: "install KubeSphere with kind",
		RunE:  opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.name, "name", "n", "kind",
		"The name of kind")
	flags.StringVarP(&opt.version, "version", "v", "v1.18.2",
		"The version for Kubernetes")
	flags.StringToStringVarP(&opt.portMappings, "portMappings", "", nil,
		"The extraPortMappings")
	flags.StringVarP(&opt.ksVersion, "ksVersion", "", "v3.0.0",
		"The version of KubeSphere")
	flags.StringSliceVarP(&opt.components, "components", "", []string{},
		"Which components will enable")
	flags.BoolVarP(&opt.Reset, "reset", "", false, "")
	flags.StringVarP(&opt.Nightly, "nightly", "", "", "")
	return
}

func (o *kindOption) reset(cmd *cobra.Command, args []string) (err error) {
	var tag string
	// try to parse the nightly date
	if o.Nightly == "latest" {
		tag = fmt.Sprintf("nightly-%s", time.Now().AddDate(0, 0, -1).Format("20060102"))
	} else if o.Nightly != "" {
		layout := "2006-01-02"
		var targetDate time.Time
		if targetDate, err = time.Parse(layout, o.Nightly); err == nil {
			tag = fmt.Sprintf("nightly-%s", targetDate.Format("20060102"))
		}
	}

	if err = pullAndLoadImage(fmt.Sprintf("kubespheredev/ks-installer:%s", tag)); err != nil {
		return
	}
	if err = pullAndLoadImage(fmt.Sprintf("kubespheredev/ks-apiserver:%s", tag)); err != nil {
		return
	}
	if err = pullAndLoadImage(fmt.Sprintf("kubespheredev/ks-controller-manager:%s", tag)); err != nil {
		return
	}
	if err = pullAndLoadImage(fmt.Sprintf("kubespheredev/ks-console:%s", tag)); err != nil {
		return
	}
	if err = execCommand("kubectl", "ks", "reset", "com", "--nightly", "latest", "-a"); err != nil {
		return
	}
	return
}

func (o *kindOption) runE(cmd *cobra.Command, args []string) (err error) {
	if o.Reset {
		return o.reset(cmd, args)
	}

	writeConfigFile("config.yaml", o.portMappings)
	if err = execCommand("kind", "create", "cluster", "--image", "kindest/node:v1.18.2", "--config", "config.yaml", "--name", o.name); err != nil {
		return
	}

	if err = execCommand("kubectl", "cluster-info", " --context", fmt.Sprintf("kind-%s", o.name)); err != nil {
		return
	}

	if err = loadCoreImageOfKS(); err != nil {
		return
	}

	if err = execCommand("kubectl", "apply", "-f", fmt.Sprintf("https://github.com/kubesphere/ks-installer/releases/download/%s/kubesphere-installer.yaml", o.ksVersion)); err != nil {
		return
	}

	if err = execCommand("kubectl", "apply", "-f", fmt.Sprintf("https://github.com/kubesphere/ks-installer/releases/download/%s/cluster-configuration.yaml", o.ksVersion)); err != nil {
		return
	}

	fmt.Println(`kubectl -n kubesphere-system patch deploy ks-installer -type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"IfNotPresent"}]'`)

	for _, com := range o.components {
		switch com {
		case "devops":
			err = loadImagesOfDevOps()
		}
	}
	return
}

func loadImagesOfDevOps() (err error) {
	if err = pullAndLoadImage("kubesphere/jenkins-uc:v3.0.0"); err != nil {
		return
	}
	if err = pullAndLoadImage("jenkins/jenkins:2.176.2"); err != nil {
		return
	}
	if err = pullAndLoadImage("jenkins/jnlp-slave:3.27-1"); err != nil {
		return
	}
	if err = pullAndLoadImage("kubesphere/builder-base:v2.1.0"); err != nil {
		return
	}
	if err = pullAndLoadImage("kubesphere/builder-nodejs:v2.1.0"); err != nil {
		return
	}
	if err = pullAndLoadImage("kubesphere/builder-go:v2.1.0"); err != nil {
		return
	}
	if err = pullAndLoadImage("kubesphere/builder-maven:v2.1.0"); err != nil {
		return
	}
	return
}

func loadCoreImageOfKS() (err error) {
	if err = pullAndLoadImage("kubesphere/ks-installer:v3.0.0"); err != nil {
		return
	}
	if err = pullAndLoadImage("kubesphere/ks-apiserver:v3.0.0"); err != nil {
		return
	}
	if err = pullAndLoadImage("kubesphere/ks-controller-manager:v3.0.0"); err != nil {
		return
	}
	if err = pullAndLoadImage("kubesphere/ks-console:v3.0.0"); err != nil {
		return
	}
	if err = pullAndLoadImage("redis:5.0.5-alpine"); err != nil {
		return
	}
	if err = pullAndLoadImage("osixia/openldap:1.3.0"); err != nil {
		return
	}
	if err = pullAndLoadImage("minio/minio:RELEASE.2019-08-07T01-59-21Z"); err != nil {
		return
	}
	if err = pullAndLoadImage("mysql:8.0.11"); err != nil {
		return
	}
	return
}

func pullAndLoadImage(image string) (err error) {
	if err = execCommand("docker", "pull", image); err == nil {
		err = execCommand("kind", "load", "docker-image", image)
	}
	return
}

func writeConfigFile(filename string, portMapping map[string]string) {
	tpl := template.New("config")
	temp, err := tpl.Parse(`
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
	fmt.Println(err)
	f, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0600)
	fmt.Println(portMapping)
	if err := temp.Execute(f, portMapping); err != nil {
		fmt.Println(err)
	}
}

type kindOption struct {
	name         string
	version      string
	portMappings map[string]string
	ksVersion    string
	components   []string

	Reset   bool
	Nightly string
}

func execCommand(name string, arg ...string) (err error) {
	command := exec.Command(name, arg...)

	//var stdout []byte
	//var errStdout error
	stdoutIn, _ := command.StdoutPipe()
	stderrIn, _ := command.StderrPipe()
	err = command.Start()
	if err != nil {
		return err
	}

	// cmd.Wait() should be called only after we finish reading
	// from stdoutIn and stderrIn.
	// wg ensures that we finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_, _ = copyAndCapture(os.Stdout, stdoutIn)
		wg.Done()
	}()

	_, _ = copyAndCapture(os.Stderr, stderrIn)

	wg.Wait()

	err = command.Wait()
	return
}

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}
