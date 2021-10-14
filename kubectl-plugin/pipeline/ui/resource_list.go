package ui

import (
	"context"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/ui/dialog"
	"github.com/rivo/tview"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

// ResourceList represents a list of a Kubernetes resource
type ResourceList struct {
	*tview.List

	//client *rest.RESTClient
	client dynamic.Interface
	app    *tview.Application
	stack  *Stack

	// inner fields
	itemAddingListeners []ItemAddingListener
	watch               watch.Interface
	resource            schema.GroupVersionResource
}

// NewResourceList creates a list for Kubernetes resource
func NewResourceList(client dynamic.Interface, app *tview.Application, stack *Stack) *ResourceList {
	list := tview.NewList()
	list.SetBorder(true)

	resourceList := &ResourceList{
		List:   list,
		client: client,
		app:    app,
		stack:  stack,
	}
	list.SetInputCapture(resourceList.eventHandler)
	return resourceList
}

// Load loads the data
func (r *ResourceList) Load(ns string, resource schema.GroupVersionResource, labelSelector string) {
	r.Stop().Clear()
	r.SetTitle("loading")
	r.resource = resource

	go func() {
		var err error
		if r.watch, err = r.client.Resource(r.resource).Namespace(ns).Watch(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		}); err == nil {
			for event := range r.watch.ResultChan() {
				switch event.Type {
				case watch.Added:
					unss := event.Object.(*unstructured.Unstructured)
					r.AddItem(unss.GetName(), "", 0, nil)
					r.callItemAddingListener(unss.GetName())
				case watch.Deleted:
					for i := 0; i < r.GetItemCount(); i++ {
						name, _ := r.GetItemText(i)
						unss := event.Object.(*unstructured.Unstructured)
						ss := &corev1.Namespace{}
						if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unss.Object, ss); err == nil {
							if name == ss.Name {
								r.RemoveItem(i)
								break
							}
						}
					}
				}
				r.SetTitle(fmt.Sprintf("%s[%d]", resource.Resource, r.GetItemCount()))
				r.app.Draw()
			}
		}
	}()
}

func (r *ResourceList) eventHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlD:
		name, _ := r.GetItemText(r.GetCurrentItem())
		r.stack.Push(dialog.ShowDelete(fmt.Sprintf("Delete %s [%s]?", r.resource.Resource, name), func() {
			r.stack.Pop()
			r.deleteItemAndResource(name)
		}, func() {
			r.stack.Pop()
		}))
	}
	return event
}

func (r *ResourceList) deleteItemAndResource(name string) {
	_ = r.client.Resource(r.resource).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func (r *ResourceList) callItemAddingListener(name string) {
	for _, listener := range r.itemAddingListeners {
		listener(name)
	}
}

// PutItemAddingListener puts an ItemAddingListener
func (r *ResourceList) PutItemAddingListener(listener ItemAddingListener) {
	r.itemAddingListeners = append(r.itemAddingListeners, listener)
}

// Stop stops reload the data
func (r *ResourceList) Stop() *ResourceList {
	if r.watch != nil {
		r.watch.Stop()
		r.watch = nil
	}
	return r
}

// ItemAddingListener callback when adding an item
type ItemAddingListener func(string)
