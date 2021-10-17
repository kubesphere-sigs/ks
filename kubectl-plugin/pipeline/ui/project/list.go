package project

import (
	"context"
	"github.com/gdamore/tcell/v2"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/ui"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/ui/dialog"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/rivo/tview"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

const (
	labelSelector = "kubesphere.io/devopsproject"
)

// List is the list view of DevOps project
type List struct {
	*ui.ResourceList

	client       dynamic.Interface
	stack        *ui.Stack
	app          *tview.Application
	inputCapture func(event *tcell.EventKey) *tcell.EventKey
}

// NewProjectList creates a project list view
func NewProjectList(client dynamic.Interface, app *tview.Application, stack *ui.Stack) (list *List) {
	list = &List{
		ResourceList: ui.NewResourceList(client, app, stack),
		client:       client,
		stack:        stack,
		app:          app,
	}
	list.Load("", types.GetNamespaceSchema(), labelSelector)

	list.inputCapture = list.GetInputCapture()
	list.SetInputCapture(list.eventHandler)

	go func() {
		// let users choose if create a project if it's empty
		list.detectIfNoProjects()
	}()
	return
}

func (l *List) detectIfNoProjects() {
	if unstructuredList, err := l.client.Resource(types.GetNamespaceSchema()).
		List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector}); err == nil && len(unstructuredList.Items) == 0 {
		l.stack.Push(dialog.ShowConfirm("No DevOps projects found, if you want to create one?", func() {
			l.stack.Pop()
			l.showProjectCreatingForm()
		}, func() {
			l.stack.Pop()
		}))
		l.app.Draw()
	}
}

func (l *List) eventHandler(event *tcell.EventKey) *tcell.EventKey {
	switch key := event.Rune(); key {
	case 'p':
		l.showProjectCreatingForm()
	}
	return l.inputCapture(event)
}

func (l *List) showProjectCreatingForm() {
	form := NewProjectForm(l.client)
	form.SetConfirmEvent(func() {
		l.stack.Pop()
	}).SetCancelEvent(func() {
		l.stack.Pop()
	})
	l.stack.Push(form)
}
