package types

import "k8s.io/apimachinery/pkg/runtime/schema"

// GetUserSchema returns the schema of users
func GetUserSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}
}

// GetPipelineSchema returns the schema of pipelines
func GetPipelineSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "devops.kubesphere.io",
		Version:  "v1alpha3",
		Resource: "pipelines",
	}
}

// GetNamespaceSchema returns the schema of namespaces
func GetNamespaceSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  "v1",
		Resource: "namespaces",
	}
}

// GetDeploySchema returns the schema of deploy
func GetDeploySchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
}

// GetClusterConfiguration returns the schema of ClusterConfiguration
func GetClusterConfiguration() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "installer.kubesphere.io",
		Version:  "v1alpha1",
		Resource: "clusterconfigurations",
	}
}

// GetServiceSchema returns the schema of service
func GetServiceSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  "v1",
		Resource: "services",
	}
}
