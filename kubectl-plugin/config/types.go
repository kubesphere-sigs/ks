package config

import "k8s.io/client-go/dynamic"

type clusterOption struct {
	role      string
	jwtSecret string

	// inner fields
	client dynamic.Interface
}

type migrateOption struct {
	service   string
	devops    bool
	namespace string

	// inner fields
	client dynamic.Interface
}

type kubeSphereConfig struct {
	Authentication authentication
}

type authentication struct {
	JwtSecret string
}
