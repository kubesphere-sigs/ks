package ui

import "github.com/rivo/tview"

// Stack is the stack of the views
type Stack struct {
	views []tview.Primitive
	count int
	app   *tview.Application

	listeners []func()
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
	s.fireChanges()
}

// Pop pops a view and show the preview one
func (s *Stack) Pop() tview.Primitive {
	if s.count == 0 {
		return nil
	}
	s.count--
	s.showCurrentView()
	s.fireChanges()
	return s.views[s.count]
}

// GetCurrent returns the current view
func (s *Stack) GetCurrent() tview.Primitive {
	if s.count < 1 {
		return nil
	}
	return s.views[s.count-1]
}

// AddChangeListener adds a change listener
func (s *Stack) AddChangeListener(listener func()) {
	s.listeners = append(s.listeners, listener)
}

func (s *Stack) fireChanges() {
	for _, listener := range s.listeners {
		listener()
	}
}

func (s *Stack) showCurrentView() {
	if current := s.GetCurrent(); current != nil {
		s.app.SetRoot(current, true)
	}
}
