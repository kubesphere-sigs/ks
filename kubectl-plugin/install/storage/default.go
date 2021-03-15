package storage

import (
	"context"
	kstypes "github.com/linuxsuren/ks/kubectl-plugin/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

// HasDefaultStorageClass to check if there's a default storageClass in current cluster
func HasDefaultStorageClass(client dynamic.Interface) (hasDefault bool) {
	ctx := context.TODO()
	if list, err := client.Resource(kstypes.GetStorageClassSchema()).List(ctx, metav1.ListOptions{}); err == nil {
		for _, item := range list.Items {
			if val, ok := item.GetAnnotations()["storageclass.beta.kubernetes.io/is-default-class"]; ok {
				hasDefault = val == "true"
				return
			}
		}
	}
	return
}
