package registry

import (
	"context"
	"fmt"
	"github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

// NewRegistryCmd returns a command of pipeline
func NewRegistryCmd(client dynamic.Interface) (cmd *cobra.Command) {
	ctx := context.TODO()

	cmd = &cobra.Command{
		Use:     "registry",
		Aliases: []string{"reg"},
		Short:   "start a registry locally",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			obj := &unstructured.Unstructured{}
			content := getRegistryDeploy()

			client.Resource(types.GetDeploySchema()).Namespace("kubesphere-system").Delete(ctx, "registry", metav1.DeleteOptions{})
			client.Resource(types.GetServiceSchema()).Namespace("kubesphere-system").Delete(ctx, "registry", metav1.DeleteOptions{})

			if err = yaml.Unmarshal([]byte(content), obj); err == nil {
				if _, err = client.Resource(types.GetDeploySchema()).Namespace("kubesphere-system").Create(ctx, obj, metav1.CreateOptions{}); err != nil {
					err = fmt.Errorf("failed when create deploy, %#v", err)
					return
				}
			}

			svcContent := getService()
			if err = yaml.Unmarshal([]byte(svcContent), obj); err == nil {
				fmt.Println(obj)
				if _, err = client.Resource(types.GetServiceSchema()).Namespace("kubesphere-system").Create(ctx, obj, metav1.CreateOptions{}); err != nil {
					err = fmt.Errorf("failed when create service, %#v", err)
				}
			}
			return
		},
	}
	return
}

func getService() string {
	return `
apiVersion: v1
kind: Service
metadata:
  name: registry
  namespace: kubesphere-system
spec:
  ports:
  - name: registry
    nodePort: 32000
    port: 5000
    protocol: TCP
    targetPort: 5000
  selector:
    app: registry
  type: NodePort
`
}

func getRegistryDeploy() string {
	return `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: registry
  namespace: kubesphere-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: registry
  template:
    metadata:
      labels:
        app: registry
    spec:
      containers:
      - image: registry:2
        imagePullPolicy: IfNotPresent
        name: registry
        ports:
        - containerPort: 5000
          protocol: TCP
`
}
