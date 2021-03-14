package common

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

// GetClient returns the k8s client
func GetClient() (client dynamic.Interface, clientSet *kubernetes.Clientset, err error) {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	var config *rest.Config

	if config, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err == nil {
		client, err = dynamic.NewForConfig(config)
		clientSet, err = kubernetes.NewForConfig(config)
	}
	return
}
