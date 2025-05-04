package tui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/jmoiron/sqlx"
	"github.com/rivo/tview"
	"github.com/maccalsa/bashhub/internal/database"
)

type UI struct {
	app     *tview.Application
	db      *sqlx.DB
	list    *tview.List
	details *tview.TextView
	root    tview.Primitive // root primitive
}

func NewUI(db *sqlx.DB) *UI {
	ui := &UI{
		app:     tview.NewApplication(),
		db:      db,
		list:    tview.NewList().ShowSecondaryText(true),
		details: tview.NewTextView().SetDynamicColors(true).SetWrap(true),
	}

	ui.list.SetBorder(true).SetTitle(" Scripts ")
	ui.details.SetBorder(true).SetTitle(" Details ")

	return ui
}

func (ui *UI) loadScripts() {
	scripts, err := database.GetScripts(ui.db)
	if err != nil {
		ui.details.SetText(fmt.Sprintf("[red]Error loading scripts: %v", err))
		return
	}

	ui.list.Clear()

	if len(scripts) == 0 {
		ui.list.AddItem("No scripts found", "", 0, nil)
		ui.details.SetText("[yellow]No scripts to display. Add scripts to the database.")
		return
	}

	for idx, script := range scripts {
		script := script // capture loop variable
		ui.list.AddItem(script.Name, script.Description, 0, func() {
			ui.details.SetText(fmt.Sprintf("[yellow]%s\n\n[white]%s", script.Name, script.Content))
		})
		if idx == 0 {
			ui.details.SetText(fmt.Sprintf("[yellow]%s\n\n[white]%s", script.Name, script.Content))
		}
	}
}

func (ui *UI) confirmDeleteScript() {
	index := ui.list.GetCurrentItem()
	if index < 0 {
		return
	}

	name, _ := ui.list.GetItemText(index)

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Delete script '%s'?", name)).
		AddButtons([]string{"Cancel", "Delete"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				scripts, _ := database.GetScripts(ui.db)
				scriptID := scripts[index].ID
				if err := database.DeleteScript(ui.db, scriptID); err != nil {
					ui.details.SetText(fmt.Sprintf("[red]Failed to delete script: %v", err))
				} else {
					ui.loadScripts()
					ui.details.SetText("[green]Script deleted.")
				}
			}
			ui.app.SetRoot(ui.root, true)
		})

	ui.app.SetRoot(modal, false)
}


func (ui *UI) showCreateForm() {
	form := tview.NewForm()

	form.AddInputField("Name", "", 20, nil, nil).
		AddInputField("Description", "", 40, nil, nil).
		AddTextArea("Content", "", 60, 10, 0, nil).
		AddButton("Save", func() {
			name := form.GetFormItemByLabel("Name").(*tview.InputField).GetText()
			description := form.GetFormItemByLabel("Description").(*tview.InputField).GetText()
			content := form.GetFormItemByLabel("Content").(*tview.TextArea).GetText()

			script := database.Script{Name: name, Description: description, Content: content}
			if err := database.CreateScript(ui.db, script); err != nil {
				ui.details.SetText(fmt.Sprintf("[red]Failed to create script: %v", err))
			} else {
				ui.loadScripts()
				ui.details.SetText("[green]Script created successfully.")
			}
			ui.app.SetRoot(ui.root, true)
		}).
		AddButton("Cancel", func() {
			ui.app.SetRoot(ui.root, true)
		})

	form.SetBorder(true).SetTitle("New Script").SetTitleAlign(tview.AlignLeft)
	ui.app.SetRoot(form, true)
}

func (ui *UI) handleKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'C', 'c':
		ui.showCreateForm()
	case 'D', 'd':
		ui.confirmDeleteScript()
	}
	return event
}

func (ui *UI) Run() error {
	ui.loadScripts()

	ui.list.SetBorder(true).SetTitle("Scripts (Press [green]C[white]=Create, [red]D[white]=Delete)")
	ui.details.SetBorder(true).SetTitle("Details")

	ui.list.SetDoneFunc(func() {
		ui.app.Stop()
	})

	ui.list.SetInputCapture(ui.handleKeys)

	ui.root = tview.NewFlex().
		AddItem(ui.list, 0, 1, true).
		AddItem(ui.details, 0, 2, false)

	return ui.app.SetRoot(ui.root, true).Run()
}