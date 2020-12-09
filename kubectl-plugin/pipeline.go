package main

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// NewPipelineCmd returns a command of pipeline
func NewPipelineCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "pipeline",
		Aliases: []string{"pip"},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
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

	cmd.AddCommand(NewDelPipelineCmd(client))
	return
}

// NewDelPipelineCmd returns a command to delete pipelines
func NewDelPipelineCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"del", "remove", "rm"},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var pips []string
			var ns string
			if ns, pips, err = getPipelinesWithConfirm(client, args); err == nil {
				for _, pip := range pips {
					fmt.Println(pip)
					if err = client.Resource(GetPipelineSchema()).Namespace(ns).Delete(context.TODO(), pip, metav1.DeleteOptions{}); err != nil {
						break
					}
				}
			}
			return
		},
	}
	return
}

func getPipelinesWithConfirm(client dynamic.Interface, args []string) (ns string, pips []string, err error) {
	var allPips []string
	if ns, allPips, err = getPipelines(client, args); err != nil {
		return
	}

	prompt := &survey.MultiSelect{
		Message: "Please select the namespace whose you want to check:",
		Options: allPips,
	}
	err = survey.AskOne(prompt, &pips)
	return
}

func getPipelines(client dynamic.Interface, args []string) (ns string, pips []string, err error) {
	if len(args) >= 2 {
		ns, pips = args[0], args[1:]
		return
	} else if len(args) == 1 {
		ns = args[0]
	} else {
		if ns, err = getNamespace(client, args); err != nil {
			return
		}
	}

	var list *unstructured.UnstructuredList
	if list, err = client.Resource(GetPipelineSchema()).Namespace(ns).List(context.TODO(), metav1.ListOptions{}); err == nil {
		for _, item := range list.Items {
			pips = append(pips, item.GetName())
		}
	}
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
			Message: "Please select the namespace whose you want to check:",
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
	if list, err := client.Resource(GetNamespaceSchema()).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "kubesphere.io/devopsproject",
	}); err == nil {
		nsList = make([]string, len(list.Items))

		for i, item := range list.Items {
			nsList[i] = item.GetName()
		}
	}
	return
}
