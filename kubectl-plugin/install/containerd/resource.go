package containerd

import (
	// Enable go embed
	_ "embed"
)

//go:embed config.toml
var configToml string

// GetConfigToml returns the default containerd config file content
func GetConfigToml() string {
	return configToml
}

//go:embed crictl.yaml
var crictl string

// GetCrictl returns the default crictl config file content
func GetCrictl() string {
	return crictl
}

//go:embed kk_config.yaml
var kkConfig string

// GetKKConfig returns the default kubekey config file content
func GetKKConfig() string {
	return kkConfig
}

//go:embed containerd.service
var containerdService string

// GetContainerdService returns the default containerd.service file content
func GetContainerdService() string {
	return containerdService
}
