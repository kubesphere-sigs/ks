package ui

import "github.com/rivo/tview"

// Stack is the stack of the views
type Stack struct {
	views []tview.Primitive
	count int
	app   *tview.Application
}

// NewStack creates a new stack instance
func NewStack(app *tview.Application) *Stack {
	return &Stack{app: app}
}

// Push pushes a view and show it
func (s *Stack) Push(view tview.Primitive) {
	s.views = append(s.views[:s.count], view)
	s.count++
	s.showCurrentView()
}

// Pop pops a view and show the preview one
func (s *Stack) Pop() tview.Primitive {
	if s.count == 0 {
		return nil
	}
	s.count--
	s.showCurrentView()
	return s.views[s.count]
}

// GetCurrent returns the current view
func (s *Stack) GetCurrent() tview.Primitive {
	if s.count < 1 {
		return nil
	}
	return s.views[s.count-1]
}

func (s *Stack) showCurrentView() {
	if current := s.GetCurrent(); current != nil {
		s.app.SetRoot(current, true)
	}
}
