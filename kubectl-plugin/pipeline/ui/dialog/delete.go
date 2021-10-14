package dialog

import "github.com/rivo/tview"

type (
	okFunc     func()
	cancelFunc func()
)

// ShowDelete pops a resource deletion dialog.
func ShowDelete(msg string, ok okFunc, cancel cancelFunc) *tview.Modal {
	confirm := tview.NewModal()
	confirm.SetText(msg)
	confirm.AddButtons([]string{"OK", "Cancel"})
	confirm.SetDoneFunc(func(_ int, label string) {
		switch label {
		case "OK":
			ok()
		}
		cancel()
	})
	return confirm
}
