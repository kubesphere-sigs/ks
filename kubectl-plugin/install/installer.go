package install

import (
	"fmt"
	"github.com/linuxsuren/ks/kubectl-plugin/common"
	"github.com/linuxsuren/ks/kubectl-plugin/install/storage"
	"github.com/spf13/cobra"
	"html/template"
	"k8s.io/client-go/dynamic"
	"os"
	"path"
)

func newInstallerCmd() (cmd *cobra.Command) {
	opt := &installerOption{}

	cmd = &cobra.Command{
		Use:   "installer",
		Short: "Install KubeSphere in an exist Kubernetes with ks-installer",
		Long: `Install KubeSphere in an exist Kubernetes with ks-installer
You can get more details about the ks-installer from https://github.com/kubesphere/ks-installer`,
		Example: "ks install installer --nightly latest --components DevOps,Logging",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.version, "version", "", "v3.0.0",
		"The version of KubeSphere which you want to install")
	flags.StringVarP(&opt.nightly, "nightly", "", "",
		"The nightly version you want to install")
	flags.StringArrayVarP(&opt.components, "components", "", []string{},
		"The components that you want to Enabled with KubeSphere")
	return
}

type installerOption struct {
	version    string
	nightly    string
	components []string

	// inner fields
	client      dynamic.Interface
	ksInstaller common.KSInstallerSpec
}

func (o *installerOption) preRunE(_ *cobra.Command, args []string) (err error) {
	if o.client, _, err = common.GetClient(); err != nil {
		err = fmt.Errorf("unable to init the k8s client, error: %v", err)
	}

	// check if ks was exists

	_, o.nightly = common.GetNightlyTag(o.nightly)

	// parse the ks-installer
	o.ksInstaller = common.KSInstallerSpec{
		Version: o.version,
	}
	if o.nightly != "" {
		o.ksInstaller.ImageNamespace = "kubespheredev"
		o.ksInstaller.Version = o.nightly
	}
	for _, item := range o.components {
		switch item {
		case "servicemesh":
			o.ksInstaller.Servicemesh.Enabled = true
		case "openpitrix":
			o.ksInstaller.Openpitrix.Enabled = true
		case "notification":
			o.ksInstaller.Notification.Enabled = true
		case "networkPolicy":
			o.ksInstaller.NetworkPolicy.Enabled = true
		case "metricsServer":
			o.ksInstaller.MetricsServer.Enabled = true
		case "logging":
			o.ksInstaller.Logging.Enabled = true
		case "events":
			o.ksInstaller.Events.Enabled = true
		case "devops":
			o.ksInstaller.DevOps.Enabled = true
		case "auditing":
			o.ksInstaller.Auditing.Enabled = true
		case "alerting":
			o.ksInstaller.Alerting.Enabled = true
		}
	}

	if !storage.HasDefaultStorageClass(o.client) {
		err = storage.CreateEBSAsDefault()
	}
	return
}

func (o *installerOption) getCrdAndCC() (crd, cc string, err error) {
	var crdTmp *template.Template
	if crdTmp, err = template.New("crd").Parse(ks3_0CRD); err != nil {
		err = fmt.Errorf("failed to parse the crd template, error: %v", err)
		return
	}
	var ccTmp *template.Template
	if ccTmp, err = template.New("cc").Parse(clusterConfiguration); err != nil {
		err = fmt.Errorf("failed to parse the clusterConfigration template, error: %v", err)
		return
	}

	crd = path.Join(os.TempDir(), "crd.yaml")
	cc = path.Join(os.TempDir(), "cc.yaml")

	var crdOut, ccOut *os.File
	if crdOut, err = os.Create(crd); err != nil {
		return
	}
	if ccOut, err = os.Create(cc); err != nil {
		return
	}

	if err = crdTmp.Execute(crdOut, o.ksInstaller); err != nil {
		return
	}
	err = ccTmp.Execute(ccOut, o.ksInstaller)
	return
}

func (o *installerOption) runE(_ *cobra.Command, args []string) (err error) {
	var crdPath, ccPath string
	if crdPath, ccPath, err = o.getCrdAndCC(); err != nil {
		return
	}

	defer func() {
		// clean the temporary files
		_ = os.RemoveAll(crdPath)
		_ = os.RemoveAll(ccPath)
	}()

	commander := Commander{}
	if err = commander.execCommand("kubectl", "apply", "-f", crdPath); err == nil {
		err = commander.execCommand("kubectl", "apply", "-f", ccPath)
	}
	return
}

var localStorageClass = `
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast
provisioner: kubernetes.io/storageos
parameters:
  pool: default
  description: Kubernetes volume
  fsType: ext4
  adminSecretNamespace: default
  adminSecretName: storageos-secret
`
