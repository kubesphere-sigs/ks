package common

// GetPluginAbleComponents returns the component list which can plug-in or plug-out
func GetPluginAbleComponents() []string {
	return []string{
		"devops", "alerting", "auditing", "events", "logging", "metrics_server", "networkpolicy",
		"notification", "openpitrix", "servicemesh",
	}
}
