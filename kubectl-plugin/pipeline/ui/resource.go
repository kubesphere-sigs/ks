package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/ui/actions"
	"github.com/rivo/tview"
)

// Resource represents a resource view
type Resource interface {
	tview.Primitive

	GetKeyActions() actions.KeyActions
	AddKeyActions(actions.KeyActions)

	SetKind(string)
	GetKind() string
}

// ResourcePrimitive represents a grid view
type ResourcePrimitive struct {
	*tview.Grid
}

// GetKeyActions returns the key actions
func (r *ResourcePrimitive) GetKeyActions() actions.KeyActions {
	return map[tcell.Key]actions.KeyAction{
		'r': {
			Description: "Run goto PipelineRun",
		},
		'R': {
			Description: "Run Pipeline",
		},
		'v': {
			Description: "List the PipelineRuns",
		},
		'c': {
			Description: "Create a Pipeline",
		},
		'y': {
			Description: "View the as YAML",
		},
	}
}

// AddKeyActions add key actions
func (r *ResourcePrimitive) AddKeyActions(actions.KeyActions) {
}

// SetKind set the kind
func (r *ResourcePrimitive) SetKind(string) {
}

// GetKind returns the kind
func (r *ResourcePrimitive) GetKind() string {
	return ""
}
