package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/gdamore/tcell/v2"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/option"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/tpl"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/ui"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/ui/project"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strings"
)

type dashboardOption struct {
	client     dynamic.Interface
	clientset  *kubernetes.Clientset
	restClient *rest.RESTClient

	namespace             string
	pipeline              string
	namespaceWorkspaceMap map[string]string
	namespaceProjectMap   map[string]string

	stack            *ui.Stack
	header           *ui.Header
	footer           *tview.Table
	app              *tview.Application
	pipelineListView *ui.ResourceTable
}

func newDashboardCmd() (cmd *cobra.Command) {
	opt := &dashboardOption{
		namespaceWorkspaceMap: map[string]string{},
		namespaceProjectMap:   map[string]string{},
	}
	cmd = &cobra.Command{
		Use:     "dashboard",
		Aliases: []string{"dash"},
		RunE:    opt.runE,
	}
	return
}

func (o *dashboardOption) runE(cmd *cobra.Command, args []string) (err error) {
	o.app = tview.NewApplication()
	o.stack = ui.NewStack(o.app)

	o.client = common.GetDynamicClient(cmd.Root().Context())
	o.clientset = common.GetClientset(cmd.Root().Context())
	o.restClient = common.GetRestClient(cmd.Root().Context())

	grid := ui.ResourcePrimitive{
		Grid: tview.NewGrid(),
	}
	grid.SetRows(3, 0, 3)
	grid.SetColumns(30, 0, 30)
	grid.SetBorder(true)
	o.header = ui.NewHeader(o.clientset, o.stack)
	o.footer = tview.NewTable()
	grid.AddItem(o.header, 0, 0, 1, 3, 0, 0, false)
	grid.AddItem(o.footer, 2, 0, 1, 3, 0, 0, false)
	grid.AddItem(o.createNamespaceList(), 1, 0, 1, 1, 0, 100, true)
	grid.AddItem(o.createPipelineList(), 1, 1, 1, 2, 0, 100, false)
	go func() {
		o.getComponentsInfo()
	}()
	o.stack.Push(grid)
	if err = o.app.
		Run(); err != nil {
		panic(err)
	}
	return
}

func (o *dashboardOption) getComponentsInfo() {
	// TODO consider reading the namespace from somewhere
	if watchEvent, err := o.clientset.AppsV1().Deployments("kubesphere-devops-system").Watch(context.TODO(), metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/instance=devops",
	}); err == nil {
		for event := range watchEvent.ResultChan() {
			deploy := event.Object.(*v1.Deployment)
			updateTable(o.footer, deploy.Name, deploy.Name,
				fmt.Sprintf("%d/%d", deploy.Status.ReadyReplicas, deploy.Status.Replicas), deploy.Spec.Template.Spec.Containers[0].Image)
		}
	}
}

func updateTable(table *tview.Table, name string, values ...string) {
	rowCount := table.GetRowCount()
	found := false
	for i := 0; i < rowCount; i++ {
		cell := table.GetCell(i, 0)
		if cell.Text == name {
			for j, val := range values {
				table.SetCellSimple(i, j, val)
			}
			found = true
			break
		}
	}

	if !found {
		rowCount++
		for j, val := range values {
			table.SetCellSimple(rowCount, j, val)
		}
	}
}

func (o *dashboardOption) createPipelineList() (listView tview.Primitive) {
	table := ui.NewResourceTable(o.restClient, o.app, o.stack)
	o.pipelineListView = table
	listView = table
	oldInputCapture := table.GetInputCapture()
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch key := event.Rune(); key {
		case 'v':
			o.listPipelineRuns(0, o.namespace, "", 0)
		case 'r':
			run := &pipelineRunOpt{
				client: o.client,
			}
			row, col := table.GetSelection()
			cell := table.GetCell(row, col)
			pipeline := cell.Text
			_ = run.triggerPipeline(o.namespace, pipeline, nil)
			o.listPipelineRuns(0, o.namespace, "", 0)
		case 'R':
			run := &pipelineRunOpt{
				client: o.client,
			}
			row, col := table.GetSelection()
			cell := table.GetCell(row, col)
			pipeline := cell.Text
			_ = run.triggerPipeline(o.namespace, pipeline, nil)
		case 'c':
			o.pipelineCreationForm()
		case 'y':
			row, col := table.GetSelection()
			cell := table.GetCell(row, col)
			pipeline := cell.Text
			if strings.HasPrefix(table.GetTitle(), "pipelinerun") {
				o.resourceYAMLView(types.GetPipelineRunSchema(), o.namespace, pipeline)
			} else if strings.HasPrefix(table.GetTitle(), "pipeline") {
				o.resourceYAMLView(types.GetPipelineSchema(), o.namespace, pipeline)
			}
		}
		if event.Key() == tcell.KeyESC {
			o.listPipelines(0, o.namespace, "", 0)
		}
		return oldInputCapture(event)
	})
	return
}

func (o *dashboardOption) listPipelineRuns(index int, mainText string, secondaryText string, shortcut rune) {
	o.pipelineListView.Clear()
	o.pipelineListView.SetTitle("PipelineRuns")
	_ = o.getTable(mainText, "pipelineruns", o.pipelineListView)
}

func (o *dashboardOption) getTable(ns, kind string, table *ui.ResourceTable) (err error) {
	var labelSelector string
	if kind == "pipelineruns" {
		labelSelector = fmt.Sprintf("devops.kubesphere.io/pipeline=%s", o.pipeline)
	}
	table.Load(ns, kind, labelSelector)
	return
}

func (o *dashboardOption) resourceYAMLView(groupVer schema.GroupVersionResource, ns, name string) {
	textView := tview.NewTextView()
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			o.stack.Pop()
		}
		return event
	})
	textView.SetBorder(true)
	textView.SetTitle(fmt.Sprintf("%s(%s/%s)", groupVer.Resource, ns, name))
	var data *unstructured.Unstructured
	var err error
	if data, err = o.client.Resource(groupVer).Namespace(ns).Get(context.TODO(), name, metav1.GetOptions{}); err == nil {
		data.SetManagedFields(nil)

		buffer := bytes.NewBuffer([]byte{})
		printer := &printers.YAMLPrinter{}
		if err = printer.PrintObj(data.DeepCopyObject(), buffer); err == nil {
			textView.SetText(buffer.String())
		}
	}
	o.stack.Push(textView)
}

func (o *dashboardOption) pipelineCreationForm() {
	form := tview.NewForm()
	form.AddButton("OK", func() {
		nameItem := form.GetFormItemByLabel("Name")
		templateItem := form.GetFormItemByLabel("Template")
		if nameItem != nil && templateItem != nil {
			nameField := nameItem.(*tview.InputField)
			templateField := templateItem.(*tview.DropDown)
			_, templateName := templateField.GetCurrentOption()

			opt := &option.PipelineCreateOption{
				Name:      nameField.GetText(),
				Project:   o.namespaceProjectMap[o.namespace],
				Template:  templateName,
				Workspace: o.namespaceWorkspaceMap[o.namespace],
				Batch:     true,
				Type:      "pipeline",
				Client:    o.client,
			}
			_ = opt.ParseTemplate()
			_ = opt.CreatePipeline() // need to find a way to show the errors
		}

		o.stack.Pop()
	}).
		AddButton("Cancel", func() {
			o.stack.Pop()
		})
	form.AddDropDown("Template", tpl.GetAllTemplates(), 0, func(option string, optionIndex int) {
		if formItem := form.GetFormItemByLabel("Name"); formItem != nil {
			inputField := formItem.(*tview.InputField)
			inputField.SetText(strings.ToLower(fmt.Sprintf("%s-%s", option, randomdata.SillyName())))
		}
	})
	form.AddInputField("Name", "", 20, nil, nil)
	form.SetTitle("Create a new Pipeline").SetBorder(true)
	o.stack.Push(form)
}

func (o *dashboardOption) listPipelines(index int, mainText string, secondaryText string, shortcut rune) {
	o.pipelineListView.Clear()
	o.namespace = mainText
	o.pipelineListView.SetTitle("Pipelines")
	_ = o.getTable(mainText, "pipelines", o.pipelineListView)
	o.pipelineListView.SetSelectionChangedFunc(func(row, column int) {
		if row == 0 {
			o.pipelineListView.Select(1, 0)
		}
		cell := o.pipelineListView.GetCell(row, column)
		o.pipeline = cell.Text
	})
}

func (o *dashboardOption) createNamespaceList() (listView tview.Primitive) {
	list := project.NewProjectList(o.client, o.app, o.stack)
	list.PutItemAddingListener(func(name string) {
		if devopsProject, err := o.client.Resource(types.GetDevOpsProjectSchema()).
			Get(context.TODO(), name, metav1.GetOptions{}); err == nil {
			o.namespaceWorkspaceMap[name] = devopsProject.GetLabels()["kubesphere.io/workspace"]
			o.namespaceProjectMap[name] = devopsProject.GetGenerateName()
		}
	})
	list.SetChangedFunc(o.listPipelines)
	o.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch key := event.Rune(); key {
		case 'j':
			event = tcell.NewEventKey(tcell.KeyDown, key, tcell.ModNone)
		case 'k':
			event = tcell.NewEventKey(tcell.KeyUp, key, tcell.ModNone)
		case 'l':
			o.app.SetFocus(o.pipelineListView)
		case 'h':
			o.app.SetFocus(list)
		}
		return event
	})
	o.app.SetFocus(list)
	listView = list
	return
}
