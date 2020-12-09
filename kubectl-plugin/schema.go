package main

import "k8s.io/apimachinery/pkg/runtime/schema"

func GetUserSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}
}

func GetPipelineSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "devops.kubesphere.io",
		Version:  "v1alpha3",
		Resource: "pipelines",
	}
}

func GetNamespaceSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  "v1",
		Resource: "namespaces",
	}
}
