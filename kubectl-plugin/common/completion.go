package common

import "github.com/spf13/cobra"

// CompletionFunc is the function for command completion
type CompletionFunc func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

// NoFileCompletion avoid completion with files
func NoFileCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// ArrayCompletion return a completion  which base on an array
func ArrayCompletion(array ...string) CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return array, cobra.ShellCompDirectiveNoFileComp
	}
}

// PluginAbleComponentsCompletion returns a completion function for pluginAble components
func PluginAbleComponentsCompletion() CompletionFunc {
	return ArrayCompletion(GetPluginAbleComponents()...)
}

// KubeSphereDeploymentCompletion returns a completion function for KuebSphere deployments
func KubeSphereDeploymentCompletion() CompletionFunc {
	return ArrayCompletion(GetKubeShpereDeployment()...)
}
