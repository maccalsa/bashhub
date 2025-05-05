package tui

import (
	"fmt"

	"github.com/maccalsa/bashhub/internal/database"
	"github.com/rivo/tview"
)


func (ui *UI) showEditForm() {
	node := ui.tree.GetCurrentNode()
	if node == nil {
		return
	}

	ref := node.GetReference()
	if ref == nil {
		ui.details.SetText("[red]Cannot edit category. Please select a script.")
		return
	}

	script := ref.(database.Script)

	ui.inForm = true
	form := tview.NewForm()
	scriptContent := script.Content

	form.
		AddInputField("Name", script.Name, 20, nil, nil).
		AddInputField("Description", script.Description, 40, nil, nil).
		AddInputField("Category", script.Category, 20, nil, nil).
		AddButton("Edit Content", func() {
			content, err := launchEditor(ui.app, scriptContent)
			if err != nil {
				ui.details.SetText(fmt.Sprintf("[red]Editor error: %v", err))
			} else {
				scriptContent = content
				form.GetButton(form.GetButtonCount()-3).SetLabel("Edit Content ✔️")
			}
		}).
		AddButton("Save", func() {
			name := form.GetFormItemByLabel("Name").(*tview.InputField).GetText()
			description := form.GetFormItemByLabel("Description").(*tview.InputField).GetText()
			category := form.GetFormItemByLabel("Category").(*tview.InputField).GetText()

			if scriptContent == "" {
				ui.details.SetText("[red]Script content cannot be empty. Please edit script content first.")
				return
			}

			script.Name = name
			script.Description = description
			script.Category = category
			script.Content = scriptContent
			script.Language = DetectLanguage(scriptContent)

			if err := database.UpdateScript(ui.db, script); err != nil {
				ui.details.SetText(fmt.Sprintf("[red]Failed to update script: %v", err))
			} else {
				ui.loadScripts()
				ui.details.SetText("[green]Script updated successfully.")
			}
			ui.inForm = false
			ui.app.SetRoot(ui.root, true)
		}).
		AddButton("Cancel", func() {
			ui.inForm = false
			ui.app.SetRoot(ui.root, true)
		})

	form.SetBorder(true).SetTitle("Edit Script").SetTitleAlign(tview.AlignLeft)
	ui.app.SetRoot(form, true)
}
