package tui

import (
	"fmt"

	"github.com/maccalsa/bashhub/internal/database"
	"github.com/rivo/tview"
)

func (ui *UI) showCreateForm() {
	ui.inForm = true

	var scriptContent string // temporarily store edited content here

	form := tview.NewForm()

	form.
		AddInputField("Name", "", 20, nil, nil).
		AddInputField("Description", "", 40, nil, nil).
		AddInputField("Category", "General", 20, nil, nil). // clearly added category
		AddButton("Edit Content", func() {
			// After editing completes, your TUI restores control, completely avoiding the terminal output leak.
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
				ui.inForm = false
				ui.details.SetText("[red]Script content cannot be empty. Please edit script content first.")
				return
			}

		    language := DetectLanguage(scriptContent) // automatic detection clearly here


			script := database.Script{
				Name: name, 
				Description: description, 
				Content: scriptContent,
				Category: category,
				Language: language,
			}

			if err := database.CreateScript(ui.db, script); err != nil {
				ui.details.SetText(fmt.Sprintf("[red]Failed to create script: %v", err))
			} else {
				ui.loadScripts()
				ui.details.SetText("[green]Script created successfully.")
			}
			ui.inForm = false
			ui.app.SetRoot(ui.root, true)
		}).
		AddButton("Cancel", func() {
			ui.inForm = false
			ui.app.SetRoot(ui.root, true)
		})

	form.SetBorder(true).SetTitle("New Script").SetTitleAlign(tview.AlignLeft)
	ui.app.SetRoot(form, true)
}