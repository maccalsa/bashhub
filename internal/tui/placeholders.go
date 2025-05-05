package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/maccalsa/bashhub/internal/database"
	"github.com/maccalsa/bashhub/internal/executor"
	"github.com/rivo/tview"
)


func (ui *UI) promptPlaceholderInputs(script database.Script, placeholders []string) {
	inputs := make(map[string]string)
	form := tview.NewForm()
	ui.inForm = true

	for _, ph := range placeholders {
		form.AddInputField(ph, "", 30, nil, func(text string) {
			inputs[ph] = text
		})
	}

	form.AddButton("Run", func() {
		finalScript := executor.ReplacePlaceholders(script.Content, inputs)
		ui.inForm = false
		ui.runAndDisplay(finalScript)
	}).AddButton("Cancel", func() {
		ui.inForm = false
		ui.app.SetRoot(ui.root, true)
	})

	form.SetBorder(true).SetTitle("Fill placeholders").SetTitleAlign(tview.AlignLeft)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			ui.inForm = false
			ui.app.SetRoot(ui.root, true)
		}
		return event
	})
	ui.app.SetRoot(form, true).SetFocus(form)
}