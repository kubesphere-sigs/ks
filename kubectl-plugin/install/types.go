package install

type installReport struct {
	installTool       string
	kubernetesVersion string
	kubeSphereVersion string
	installConsume    string
	beginTime         string
	endTime           string
	os                string
	arch              string
}
