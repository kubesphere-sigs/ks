package pipeline

import (
	"context"
	"fmt"
	"github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"time"
)

func newGCCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &gcOption{
		client: client,
		pipelineCreateOption: pipelineCreateOption{
			Client: client,
		},
	}
	cmd = &cobra.Command{
		Use:     "gc",
		Short:   "Garbage collector for PipelineRuns",
		Long:    "Clean all those old PipelineRuns by age or count",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.IntVarP(&opt.maxCount, "max-count", "", 10,
		"Maximum number to keep PipelineRuns")
	flags.DurationVarP(&opt.maxAge, "max-age", "", 7*24*time.Hour,
		"Maximum age to keep PipelineRuns")
	flags.StringArrayVarP(&opt.namespaces, "namespaces", "", nil,
		"Indicate namespaces do you want to clean. Clean all namespaces if it's empty")
	return
}

type gcOption struct {
	maxCount   int
	maxAge     time.Duration
	namespaces []string

	// inner fields
	client dynamic.Interface
	pipelineCreateOption
}

func (o *gcOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	if len(o.namespaces) == 0 {
		o.namespaces = getAllNamespace(o.client)
	}
	return
}

func (o *gcOption) cleanPipelineRunInNamespace(namespace string) (err error) {
	var pipelineList *unstructured.UnstructuredList
	if pipelineList, err = o.getUnstructuredListInNamespace(namespace, types.GetPipelineRunSchema()); err != nil {
		err = fmt.Errorf("failed to get PipelineRun list, error: %v", err)
		return
	}

	items := pipelineList.Items
	toDelete := len(items) - o.maxCount
	if toDelete < o.maxCount {
		return
	}

	for i := range items {
		item := items[i]

		// check remain amount
		if toDelete <= 0 {
			break
		}

		if okToDelete(item.Object, o.maxAge) {
			delErr := o.client.Resource(types.GetPipelineRunSchema()).Namespace(namespace).Delete(
				context.TODO(), item.GetName(), metav1.DeleteOptions{})
			if delErr != nil {
				fmt.Printf("failed to delete PipelineRun %s/%s, error: %v\n", item.GetName(), namespace, delErr)
			} else {
				toDelete--
				fmt.Printf("ok to delete PipelineRun %s/%s\n", item.GetName(), namespace)
			}
		}
	}
	return
}

func okToDelete(object map[string]interface{}, maxAge time.Duration) bool {
	completionTimeStr, ok, nestErr := unstructured.NestedString(object, "status", "completionTime")
	if ok && nestErr == nil {
		if completionTime, parseErr := time.Parse(time.RFC3339, completionTimeStr); parseErr == nil {
			return completionTime.Add(maxAge).Before(time.Now())
		}
	}
	return false
}

func (o *gcOption) runE(cmd *cobra.Command, args []string) error {
	cmd.Printf("starting to gc PipelineRuns in %d namespaces\n", len(o.namespaces))
	errorsNs := []string{}

	for i := range o.namespaces {
		ns := o.namespaces[i]
		if err := o.cleanPipelineRunInNamespace(ns); err != nil {
			cmd.PrintErrf("failed to clean PipelineRuns in '%s'\n", ns)
			errorsNs = append(errorsNs, ns)
		}
	}

	if len(errorsNs) > 0 {
		return fmt.Errorf("gc failed in %d namespaces: %v", len(errorsNs), errorsNs)
	}
	return nil
}
