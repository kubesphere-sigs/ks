package pipeline

import (
	"context"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type dashboardOption struct {
	client     dynamic.Interface
	clientset  *kubernetes.Clientset
	restClient *rest.RESTClient

	namespace string
	pipeline  string

	header           *tview.TextView
	footer           *tview.Table
	app              *tview.Application
	pipelineListView *tview.Table
}

func newDashboardCmd() (cmd *cobra.Command) {
	opt := &dashboardOption{}
	cmd = &cobra.Command{
		Use:     "dashboard",
		Aliases: []string{"dash"},
		RunE:    opt.runE,
	}
	return
}

func (o *dashboardOption) runE(cmd *cobra.Command, args []string) (err error) {
	o.app = tview.NewApplication()
	o.client = common.GetDynamicClient(cmd.Root().Context())
	o.clientset = common.GetClientset(cmd.Root().Context())
	o.restClient = common.GetRestClient(cmd.Root().Context())

	newPrimitive := func(text string) *tview.TextView {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}

	grid := tview.NewGrid()
	grid.SetRows(3, 0, 3)
	grid.SetColumns(30, 0, 30)
	grid.SetBorder(true)
	o.header = newPrimitive("header")
	o.footer = tview.NewTable()
	grid.AddItem(o.header, 0, 0, 1, 3, 0, 0, false)
	grid.AddItem(o.footer, 2, 0, 1, 3, 0, 0, false)
	grid.AddItem(o.createNamespaceList(), 1, 0, 1, 1, 0, 100, true)
	grid.AddItem(o.createPipelineList(), 1, 1, 1, 2, 0, 100, false)
	go func() {
		o.getComponentsInfo()
	}()
	if err = o.app.SetRoot(grid, true).Run(); err != nil {
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
	table := tview.NewTable()
	table.SetBorder(true).SetTitle("pipelines")
	table.SetSelectable(true, false).Select(1, 0).SetFixed(1, 0)
	table.SetBorderPadding(0, 0, 1, 1)
	o.pipelineListView = table
	listView = table
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		o.header.Clear()
		o.header.SetText("(r) Run a Pipeline, (v) List the PipelineRuns")
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
		}
		if event.Key() == tcell.KeyESC {
			o.listPipelines(0, o.namespace, "", 0)
		}
		return event
	})
	return
}

func (o *dashboardOption) listPipelineRuns(index int, mainText string, secondaryText string, shortcut rune) {
	o.pipelineListView.Clear()
	o.pipelineListView.SetTitle("PipelineRuns")
	_ = o.getTable(mainText, "pipelineruns", o.pipelineListView)
}

func (o *dashboardOption) getTable(ns, kind string, table *tview.Table) (err error) {
	tableData := &metav1beta1.Table{}
	table.Clear()
	table.SetTitle(fmt.Sprintf("%s(%s)[%d]", kind, ns, 0))
	listOpt := &metav1.ListOptions{}
	if kind == "pipelineruns" {
		listOpt.LabelSelector = fmt.Sprintf("devops.kubesphere.io/pipeline=%s", o.pipeline)
	}

	if err = o.restClient.Get().Namespace(ns).Resource(kind).
		VersionedParams(listOpt, metav1.ParameterCodec).
		SetHeader("Accept", "application/json;as=Table;v=v1beta1;g=meta.k8s.io").
		Do(context.TODO()).Into(tableData); err == nil {
		for i, col := range tableData.ColumnDefinitions {
			table.SetCellSimple(0, i, col.Name)
		}
		for i, row := range tableData.Rows {
			for j, cell := range row.Cells {
				table.SetCellSimple(i+1, j, fmt.Sprintf("%v", cell))
			}
		}
		table.SetTitle(fmt.Sprintf("%s(%s)[%d]", kind, ns, len(tableData.Rows)))
	}
	return
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
	list := tview.NewList()
	list.SetBorder(true).SetTitle("namespaces")
	go func() {
		if watchEvent, err := o.client.Resource(types.GetNamespaceSchema()).Watch(context.TODO(), metav1.ListOptions{
			LabelSelector: "kubesphere.io/devopsproject",
		}); err == nil {
			for event := range watchEvent.ResultChan() {
				switch event.Type {
				case watch.Added:
					unss := event.Object.(*unstructured.Unstructured)
					ss := &corev1.Namespace{}
					if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unss.Object, ss); err == nil {
						list.AddItem(ss.Name, "", 0, nil)
					}
				case watch.Deleted:
					for i := 0; i < list.GetItemCount(); i++ {
						name, _ := list.GetItemText(i)
						unss := event.Object.(*unstructured.Unstructured)
						ss := &corev1.Namespace{}
						if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unss.Object, ss); err == nil {
							if name == ss.Name {
								list.RemoveItem(i)
								break
							}
						}
					}
				}
				o.app.Draw()
			}
		}
	}()
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
