package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/jmoiron/sqlx"
	"github.com/maccalsa/bashhub/internal/database"
	"github.com/maccalsa/bashhub/internal/executor"
	"github.com/rivo/tview"
)

type UI struct {
	app     *tview.Application
	db      *sqlx.DB
	tree    *tview.TreeView
	details *tview.TextView
	root    tview.Primitive // root primitive
	footer  *tview.TextView  // clearly added footer
	inForm  bool
	searching bool
	searchBox *tview.InputField
	searchContainer *tview.Flex
}

func NewUI(db *sqlx.DB) *UI {
	ui := &UI{
		app:     tview.NewApplication(),
		db:      db,
		tree:    tview.NewTreeView(),
		details: tview.NewTextView().SetDynamicColors(true),
		searchBox: tview.NewInputField().
			SetLabel("🔍 Search: ").
			SetFieldWidth(20).
			SetFieldBackgroundColor(tcell.ColorBlack).
			SetLabelColor(tcell.ColorYellow).
			SetFieldTextColor(tcell.ColorWhite),
		searchContainer: tview.NewFlex().SetDirection(tview.FlexRow),
	}

	ui.searchBox.
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 1).
		SetBorderColor(tcell.ColorGray)

	ui.tree.SetBorder(true).SetTitle(" Scripts ")
	ui.details.SetBorder(true).SetTitle(" Details (↑/↓ Scroll) ")

	ui.footer = tview.NewTextView()
	ui.footer.SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]Tab[white]: Switch Pane | [green]C[white]: Create Script | [orange]X[white]: Edit Script | [red]D[white]: Delete Script | [blue]E[white]: Execute Script | [cyan]Ctrl+Q[white]: Quit")

	ui.footer.SetBorder(true).SetBorderColor(tcell.ColorGray)

	return ui
}


func (ui *UI) Run() error {
	ui.loadScripts()

	ui.searchBox.SetBorder(true).SetBorderColor(tcell.ColorGray)

	ui.searchContainer.Clear().
		AddItem(ui.tree, 0, 1, true)

	mainLayout := tview.NewFlex().
		AddItem(ui.searchContainer, 0, 1, true).
		AddItem(ui.details, 0, 2, false)

	ui.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainLayout, 0, 1, true).
		AddItem(ui.footer, 3, 0, false)

	ui.app.SetFocus(ui.tree)

	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ui.inForm {
			return event
		}

		if ui.searching {
			switch event.Key() {
			case tcell.KeyEsc:
				ui.searching = false
				ui.searchBox.SetText("")
				ui.searchContainer.Clear()
				ui.searchContainer.AddItem(ui.tree, 0, 1, true)
				ui.loadScripts()
				ui.app.SetFocus(ui.tree)
				return nil
			case tcell.KeyEnter:
				ui.searching = false
				ui.searchContainer.Clear()
				ui.searchContainer.AddItem(ui.tree, 0, 1, true)
				ui.app.SetFocus(ui.tree)
				return nil
			}
			return event
		}

		ui.searchBox.SetChangedFunc(func(text string) {
			ui.filterScripts(text)
		})

		switch event.Key() {
		case tcell.KeyCtrlQ:
			ui.app.Stop()
			return nil
		case tcell.KeyTab:
			if ui.app.GetFocus() == ui.tree {
				ui.app.SetFocus(ui.details)
			} else {
				ui.app.SetFocus(ui.tree)
			}
			return nil
		case tcell.KeyBacktab: // Shift+Tab clearly goes backward
			if ui.app.GetFocus() == ui.details {
				ui.app.SetFocus(ui.tree)
			} else {
				ui.app.SetFocus(ui.details)
			}
			return nil
		}

		switch event.Rune() {
		case '/':
			if !ui.searching {
				ui.searching = true
				ui.searchContainer.Clear() // clearly reset container
				ui.searchContainer.
					AddItem(ui.searchBox, 3, 0, true). // show searchBox
					AddItem(ui.tree, 0, 1, true)
				ui.app.SetFocus(ui.searchBox)
			}
			return nil
		case 'C', 'c':
			ui.showCreateForm()
			return nil
		case 'D', 'd':
			ui.confirmDeleteScript()
			return nil
		case 'E', 'e':
			ui.executeSelectedScript()
			return nil
		case 'X', 'x':
			ui.showEditForm()
			return nil
		}

		return event
	})

	return ui.app.SetRoot(ui.root, true).Run()
}



func (ui *UI) executeSelectedScript() {
	node := ui.tree.GetCurrentNode()
	if node == nil {
		ui.details.SetText("[red]No script selected.")
		return
	}

	ref := node.GetReference()
	if ref == nil {
		ui.details.SetText("[red]Please select a valid script (not a category).")
		return
	}

	script := ref.(database.Script)

	placeholders := executor.ParsePlaceholders(script.Content)
	if len(placeholders) > 0 {
		ui.promptPlaceholderInputs(script, placeholders)
	} else {
		ui.runAndDisplay(script.Content)
	}
}


func (ui *UI) promptPlaceholderInputs(script database.Script, placeholders []string) {
	inputs := make(map[string]string)
	form := tview.NewForm()

	for _, ph := range placeholders {
		form.AddInputField(ph, "", 30, nil, func(text string) {
			inputs[ph] = text
		})
	}

	form.AddButton("Run", func() {
		finalScript := executor.ReplacePlaceholders(script.Content, inputs)
		ui.runAndDisplay(finalScript)
	}).AddButton("Cancel", func() {
		ui.app.SetRoot(ui.root, true)
	})

	form.SetBorder(true).SetTitle("Fill placeholders").SetTitleAlign(tview.AlignLeft)
	ui.app.SetRoot(form, true)
}

func (ui *UI) runAndDisplay(scriptContent string) {
	outputView := tview.NewTextView()
	outputView.
		SetDynamicColors(true).
		SetScrollable(true).
		SetRegions(true).
		SetWrap(true).
		SetChangedFunc(func() {
			ui.app.QueueUpdateDraw(func() {
				outputView.ScrollToEnd()
			})
		})

	outputView.SetBorder(true).SetTitle("Execution Output (↑/↓ Scroll) | Press Q to Quit").SetBorderColor(tcell.ColorGreen)

	outputView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, col := outputView.GetScrollOffset()
		switch event.Key() {
		case tcell.KeyDown:
			outputView.ScrollTo(row+1, col)
			return nil
		case tcell.KeyUp:
			if row > 0 {
				outputView.ScrollTo(row-1, col)
			}
			return nil
		case tcell.KeyPgDn:
			outputView.ScrollTo(row+10, col)
			return nil
		case tcell.KeyPgUp:
			if row >= 10 {
				outputView.ScrollTo(row-10, col)
			} else {
				outputView.ScrollToBeginning()
			}
			return nil
		}

		if event.Rune() == 'q' || event.Rune() == 'Q' {
			ui.app.SetRoot(ui.root, true)
			return nil
		}

		return event
	})

	go func() {
		err := executor.ExecuteScript(scriptContent, func(line string) {
			ui.app.QueueUpdateDraw(func() {
				fmt.Fprintln(outputView, tview.TranslateANSI(line))
				outputView.ScrollToEnd() // Automatically scroll to the end
			})
		})
		if err != nil {
			ui.app.QueueUpdateDraw(func() {
				fmt.Fprintf(outputView, "[red]Execution failed: %v", err)
				outputView.ScrollToEnd()
			})
		}
	}()

	ui.app.SetRoot(outputView, true)
}


