package common

// KSInstaller is the installer for KubeSphere
type KSInstaller struct {
	Spec KSInstallerSpec `yaml:"spec"`
}

// KSInstallerSpec is ks-installer
type KSInstallerSpec struct {
	Version        string
	ImageNamespace string

	Servicemesh   ComponentStatus
	Openpitrix    ComponentStatus
	Notification  ComponentStatus
	NetworkPolicy ComponentStatus
	MetricsServer ComponentStatus
	Logging       ComponentStatus
	Events        ComponentStatus
	DevOps        ComponentStatus
	Auditing      ComponentStatus
	Alerting      ComponentStatus
	Multicluster  Multicluster `yaml:"multicluster"`
}

// ComponentStatus is a common status
type ComponentStatus struct {
	Enabled bool
}

// Multicluster represents multi-cluster
type Multicluster struct {
	ClusterRole string `yaml:"clusterRole"`
}
