package common

// GetPluginAbleComponents returns the component list which can plug-in or plug-out
func GetPluginAbleComponents() []string {
	return []string{
		"devops", "alerting", "auditing", "events", "logging", "metrics_server", "networkpolicy",
		"notification", "openpitrix", "servicemesh", "metering",
	}
}

//GetKubeShpereDeployment returns the deployment of KubeSphere
func GetKubeShpereDeployment() []string {
	return []string{
		"apiserver", "controller", "console", "jenkins", "installer",
	}
}
