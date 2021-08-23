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
