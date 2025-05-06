package tui

import (
	"fmt"

	"github.com/maccalsa/bashhub/internal/database"
	"github.com/rivo/tview"
)

func (ui *UI) confirmDeleteScript() {
	node := ui.tree.GetCurrentNode()
	ui.inForm = true
	if node == nil {
		return
	}

	ref := node.GetReference()
	if ref == nil {
		ui.details.SetText("[red]Cannot delete category. Please select a script.")
		return
	}

	script := ref.(database.Script)

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Delete script '%s'?", script.Name)).
		AddButtons([]string{"Cancel", "Delete"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				if err := database.DeleteScript(ui.db, script.ID); err != nil {
					ui.details.SetText(fmt.Sprintf("[red]Failed to delete script: %v", err))
				} else {
					ui.loadScripts()
					ui.details.SetText("[green]Script deleted successfully.")
				}
			}
			ui.inForm = false
			ui.app.SetRoot(ui.root, true)
		})

	ui.app.SetRoot(modal, false)
}