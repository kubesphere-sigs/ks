package install

//import _ "embed"

var (
	//go:embed resources/cluster-configuration-3.0.yaml
	clusterConfiguration string
	//go:embed resources/kubesphere-installer-3.0.yaml
	ks3_0CRD string
)
