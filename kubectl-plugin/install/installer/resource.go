package installer

import (
	// Enable go embed
	_ "embed"
)

//go:embed cluster-configuration.yaml
var clusterConfig string

// GetClusterConfig returns the cluster configuration YAML
func GetClusterConfig() string {
	return clusterConfig
}

//go:embed kubesphere-installer.yaml
var ksInstaller string

// GetKSInstaller returns the kubesphere installer YAML
func GetKSInstaller() string {
	return ksInstaller
}
