package ui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/ui/actions"
	"github.com/linuxsuren/cobra-extension/version"
	"github.com/rivo/tview"
	"k8s.io/client-go/kubernetes"
)

// Header is the header of the Pipeline dashboard
type Header struct {
	*tview.Flex

	stack     *Stack
	clientset *kubernetes.Clientset
}

// NewHeader creates header view
func NewHeader(clientset *kubernetes.Clientset, stack *Stack) (header *Header) {
	header = &Header{
		clientset: clientset,
		stack:     stack,
	}

	layout := tview.NewFlex()
	layout.AddItem(header.newClusterInfo(), 0, 1, false)
	layout.AddItem(header.newHintsInfo(), 0, 3, false)
	header.Flex = layout
	return
}

func (h *Header) newClusterInfo() (table *tview.Table) {
	table = tview.NewTable()

	table.SetCellSimple(0, 0, "K8s Rev:")
	table.SetCellSimple(1, 0, "Version:")

	if serverVer, err := h.clientset.DiscoveryClient.ServerVersion(); err == nil {
		table.SetCellSimple(0, 1, serverVer.String())
	}
	table.SetCellSimple(1, 1, version.GetVersion())
	return
}

func (h *Header) newHintsInfo() (table *tview.Table) {
	table = tview.NewTable()

	h.stack.AddChangeListener(func() {
		var current interface{}
		current = h.stack.GetCurrent()
		if resource, ok := current.(ResourcePrimitive); ok {
			h.showKeys(table, resource.GetKeyActions())
		}
	})
	return
}

func (h *Header) showKeys(table *tview.Table, keys actions.KeyActions) {
	table.Clear()

	maxRows := 3
	numberOfKeys := len(keys)

	numberOfGroups := numberOfKeys / maxRows
	if numberOfGroups%maxRows > 0 {
		numberOfGroups++
	}

	var keyNames = map[tcell.Key]string{
		'r': "r",
		'R': "R",
		'v': "v",
		'y': "y",
		'c': "c",
	}
	for key, val := range tcell.KeyNames {
		keyNames[key] = val
	}

	row := 0
	group := 0
	for key, action := range keys {
		if row+1 > maxRows {
			row = 0
			group = group + 2
		}

		table.SetCellSimple(row, group, fmt.Sprintf("<%s>", keyNames[key]))
		table.SetCellSimple(row, group+1, action.Description)
		row = row + 1
	}
}
