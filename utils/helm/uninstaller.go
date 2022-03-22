package helm

import (
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// UninstallRequest is a release uninstall request
type UninstallRequest struct {
	ComponentName string
	KubeConfig    string
	Namespace     string
}

func (r *UninstallRequest) Do() error {
	cf := genericclioptions.NewConfigFlags(true)
	cf.KubeConfig = &r.KubeConfig
	cf.Namespace = &r.Namespace

	cfg := &action.Configuration{}
	// let the os.Stdout not print the details
	logFunc := func(format string, v ...interface{}) {}
	if err := cfg.Init(cf, r.Namespace, "", logFunc); err != nil {
		return err
	}

	uninstall := action.NewUninstall(cfg)
	_, err := uninstall.Run(r.ComponentName)
	return err
}
