package actions

import "github.com/gdamore/tcell/v2"

// KeyAction represents a shortcut
type KeyAction struct {
	Description string
}

// KeyActions is the collection of key actions
type KeyActions map[tcell.Key]KeyAction
