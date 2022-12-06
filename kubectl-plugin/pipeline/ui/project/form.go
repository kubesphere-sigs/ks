package project

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/option"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"os"
)

// DevOpsProjectForm represents a form to create DevOps project
type DevOpsProjectForm struct {
	*tview.Form

	eventConfirmCallback EventCallback
	eventCancelCallback  EventCallback
}

func init() {
	file, err := os.OpenFile("pipeline.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, file))
}

// NewProjectForm creates the form
func NewProjectForm(client dynamic.Interface) *DevOpsProjectForm {
	form := tview.NewForm()
	form.AddInputField("Workspace Name", "", 20, nil, nil)
	form.AddInputField("Project Name", "", 20, nil, nil)
	form.SetTitle("Create a new Project").SetBorder(true)

	projectForm := &DevOpsProjectForm{
		Form:                 form,
		eventConfirmCallback: doNothing,
		eventCancelCallback:  doNothing,
	}

	form.AddButton("OK", func() {
		wsItem := form.GetFormItemByLabel("Workspace Name")
		projectItem := form.GetFormItemByLabel("Project Name")
		if wsItem != nil && projectItem != nil {
			wsField := wsItem.(*tview.InputField)
			projectField := projectItem.(*tview.InputField)

			wsName := wsField.GetText()

			opt := &option.PipelineCreateOption{
				Project:   projectField.GetText(),
				Workspace: wsName,
				Batch:     true,
				Type:      "pipeline",
				Client:    client,
			}

			// the workspace CRD comes from KubeSphere instead of the DevOps
			// so, it's optional
			var wsID string
			if wsName != "" {
				var ws *unstructured.Unstructured
				var err error
				if ws, err = opt.CheckWorkspace(); err != nil {
					return
				}
				wsID = string(ws.GetUID())
			}
			_, _ = opt.CheckDevOpsProject(wsID)
		}

		projectForm.eventConfirmCallback()
	})
	form.AddButton("Cancel", projectForm.eventCancelCallback)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			projectForm.eventCancelCallback()
		}
		return event
	})
	return projectForm
}

// SetConfirmEvent set the callback function for confirm event
func (p *DevOpsProjectForm) SetConfirmEvent(callback EventCallback) *DevOpsProjectForm {
	p.eventConfirmCallback = callback
	return p
}

// SetCancelEvent set the callback function for cancel event
func (p *DevOpsProjectForm) SetCancelEvent(callback EventCallback) *DevOpsProjectForm {
	p.eventCancelCallback = callback
	return p
}

// EventCallback is the function for event callback
type EventCallback func()

var doNothing = func() {}
