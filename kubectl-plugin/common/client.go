package common

import (
	"context"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"fmt"
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

// GetDynamicClient gets the dynamic k8s client from context
func GetDynamicClient(ctx context.Context) (client dynamic.Interface) {
	factory := ctx.Value(ClientFactory{})
	client, _, _ = factory.(*ClientFactory).GetClient()
	return
}

// GetClientset gets the clientset of k8s
func GetClientset(ctx context.Context) (clientset *kubernetes.Clientset) {
	factory := ctx.Value(ClientFactory{})
	_, clientset, _ = factory.(*ClientFactory).GetClient()
	return
}

// GetRestClient returns the restClient of the Kubernetes
func GetRestClient(ctx context.Context) (client *rest.RESTClient) {
	factory := ctx.Value(ClientFactory{})
	_, _, client = factory.(*ClientFactory).GetClient()
	return
}

// ClientFactory is for getting k8s client
type ClientFactory struct {
	//client    dynamic.Interface
	//clientSet *kubernetes.Clientset
	context string
}

// GetClient returns k8s client
func (c *ClientFactory) GetClient() (client dynamic.Interface, clientSet *kubernetes.Clientset, restClient *rest.RESTClient) {
	KubernetesConfigFlags := genericclioptions.NewConfigFlags(false)
	if c.context != "" {
		KubernetesConfigFlags.Context = &c.context
	}
	if config, err := KubernetesConfigFlags.ToRESTConfig(); err == nil {
		client, err = dynamic.NewForConfig(config)
		if err != nil {
			fmt.Println(err)
		}
		clientSet, err = kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Println(err)
		}
		crdConfig := *config
		crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: types.GetPipelineSchema().Group,
			Version: types.GetPipelineSchema().Version}
		crdConfig.APIPath = "/apis"
		crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
		crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()
		restClient, _ = rest.UnversionedRESTClientFor(&crdConfig)
	}
	return
}

// SetContext sets the k8s context
func (c *ClientFactory) SetContext(ctx string) {
	c.context = ctx
}
