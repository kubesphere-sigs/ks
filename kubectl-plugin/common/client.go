package common

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// GetClient returns the k8s client
func GetClient() (client dynamic.Interface, clientSet *kubernetes.Clientset, err error) {
	KubernetesConfigFlags := genericclioptions.NewConfigFlags(false)
	if config, err := KubernetesConfigFlags.ToRESTConfig(); err == nil {
		client, err = dynamic.NewForConfig(config)
		clientSet, err = kubernetes.NewForConfig(config)
	}

	return
}
