package ui

import (
	"context"
	"fmt"
	"github.com/rivo/tview"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/client-go/rest"
	"time"
)

// ResourceTable represents a table of a Kubernetes resource
type ResourceTable struct {
	*tview.Table

	//client     dynamic.Interface
	client *rest.RESTClient
	app    *tview.Application

	// inner fields
	ticker *time.Ticker
}

// NewResourceTable creates a table for Kubernetes resource
func NewResourceTable(client *rest.RESTClient, app *tview.Application) *ResourceTable {
	table := tview.NewTable()
	table.SetBorder(true)
	table.SetSelectable(true, false).Select(1, 0).SetFixed(1, 0)
	table.SetBorderPadding(0, 0, 1, 1)

	return &ResourceTable{
		client: client,
		app:    app,
		Table:  table,
	}
}

const gvFmt = "application/json;as=Table;v=%s;g=%s, application/json"

// Load loads the data of a Kubernetes resource
func (t *ResourceTable) Load(ns, kind, labelSelector string) {
	t.Stop().Clear()
	t.SetTitle("loading")
	// TODO provide a way to let users set it
	t.ticker = time.NewTicker(time.Second * 2)

	go func() {
		ctx := context.TODO()
		// give it an initial data setting
		t.reload(ctx, ns, kind, labelSelector)

		for range t.ticker.C {
			t.reload(ctx, ns, kind, labelSelector)
		}
	}()
}

func (t *ResourceTable) reload(ctx context.Context, ns, kind, labelSelector string) {
	listOpt := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	t.SetTitle(fmt.Sprintf("%s(%s)[%d]", kind, ns, 0))
	tableData := &metav1beta1.Table{}
	if err := t.client.Get().Namespace(ns).Resource(kind).
		VersionedParams(&listOpt, metav1.ParameterCodec).
		SetHeader("Accept", fmt.Sprintf(gvFmt, metav1beta1.SchemeGroupVersion.Version, metav1beta1.GroupName)).
		Do(ctx).Into(tableData); err != nil {
		// TODO provide a better way to handle this error
		panic(err)
		return
	}

	for i, col := range tableData.ColumnDefinitions {
		t.SetCellSimple(0, i, col.Name)
	}
	for i, row := range tableData.Rows {
		for j, cell := range row.Cells {
			t.SetCellSimple(i+1, j, fmt.Sprintf("%v", cell))
		}
	}
	t.SetTitle(fmt.Sprintf("%s(%s)[%d]", kind, ns, len(tableData.Rows)))
	t.app.Draw()
}

// Stop stops the refresh data action
func (t *ResourceTable) Stop() *ResourceTable {
	if t.ticker != nil {
		t.ticker.Stop()
		t.ticker = nil
	}
	return t
}
