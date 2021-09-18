package pipeline

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

// NewPipelineCmd returns a command of pipeline
func NewPipelineCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "pipeline",
		Aliases: []string{"pip"},
		Short:   "Manage the Pipeline of KubeSphere DevOps",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			client := common.GetDynamicClient(cmd.Context())
			var pips []string
			if _, pips, err = getPipelines(client, args); err == nil {
				for _, pip := range pips {
					fmt.Println(pip)
				}
			}
			return
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (suggestion []string, directive cobra.ShellCompDirective) {
			suggestion = getAllNamespace(client)
			directive = cobra.ShellCompDirectiveNoFileComp
			return
		},
	}

	cmd.AddCommand(newDelPipelineCmd(client),
		newPipelineEditCmd(client),
		newPipelineViewCmd(client),
		newPipelineCreateCmd(client),
		newPipelineRunCmd(),
		newGCCmd(client))
	return
}

func getNamespace(client dynamic.Interface, args []string) (ns string, err error) {
	if len(args) == 0 {
		nsList := getAllNamespace(client)
		if len(nsList) == 0 {
			err = fmt.Errorf("no pipeline namespace found in this cluster")
			return
		}

		prompt := &survey.Select{
			Message: "Please select the namespace which you want to check:",
			Options: nsList,
		}
		if err = survey.AskOne(prompt, &ns); err != nil {
			return
		}
	} else {
		ns = args[0]
	}
	return
}

func getAllNamespace(client dynamic.Interface) (nsList []string) {
	if list, err := client.Resource(types.GetNamespaceSchema()).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "kubesphere.io/devopsproject",
	}); err == nil {
		nsList = make([]string, len(list.Items))

		for i, item := range list.Items {
			nsList[i] = item.GetName()
		}
	}
	return
}
