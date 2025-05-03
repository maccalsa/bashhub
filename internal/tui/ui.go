package tui

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rivo/tview"
	"github.com/maccalsa/bashhub/internal/database"
)

type UI struct {
	app     *tview.Application
	db      *sqlx.DB
	list    *tview.List
	details *tview.TextView
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

func (ui *UI) Run() error {
	ui.loadScripts()

	flex := tview.NewFlex().
		AddItem(ui.list, 0, 1, true).
		AddItem(ui.details, 0, 2, false)

	ui.app.SetRoot(flex, true).SetFocus(ui.list)
	return ui.app.Run()
}
