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

// GetDevOpsProjectSchema returns the schema of DevOpsProject
func GetDevOpsProjectSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "devops.kubesphere.io",
		Version:  "v1alpha3",
		Resource: "devopsprojects",
	}
}

// GetWorkspaceSchema returns the schema of workspace
func GetWorkspaceSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "tenant.kubesphere.io",
		Version:  "v1alpha1",
		Resource: "workspaces",
	}
}

// GetWorkspaceTemplate returns the schema of WorkspaceTemplate
func GetWorkspaceTemplate() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "tenant.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "workspacetemplates",
	}
}

// GetNamespaceSchema returns the schema of namespaces
func GetNamespaceSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  "v1",
		Resource: "namespaces",
	}
}

// GetPodSchema returns the schema of deploy
func GetPodSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
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

// GetConfigMapSchema returns the schema of ConfigMap
func GetConfigMapSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  "v1",
		Resource: "configmaps",
	}
}

// GetSecretSchema returns the schema of Secret
func GetSecretSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  "v1",
		Resource: "secrets",
	}
}

// GetStorageClassSchema returns the schema of StorageClass
func GetStorageClassSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "storage.k8s.io",
		Version:  "v1",
		Resource: "storageclasses",
	}
}

// GetS2iBuilderTemplateSchema returns the schema of S2iBuilderTemplate
func GetS2iBuilderTemplateSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "devops.kubesphere.io",
		Version:  "v1alpha1",
		Resource: "s2ibuildertemplates",
	}
}

// GetS2iBuilderSchema returns the schema of S2iBuilder
func GetS2iBuilderSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "devops.kubesphere.io",
		Version:  "v1alpha1",
		Resource: "s2ibuilders",
	}
}
