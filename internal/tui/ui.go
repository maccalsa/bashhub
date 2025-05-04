package tui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/jmoiron/sqlx"
	"github.com/rivo/tview"
	"github.com/maccalsa/bashhub/internal/database"
	"github.com/maccalsa/bashhub/internal/executor"
	"os"
	"os/exec"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"bytes"
)

type UI struct {
	app     *tview.Application
	db      *sqlx.DB
	list    *tview.List
	details *tview.TextView
	root    tview.Primitive // root primitive
	footer  *tview.TextView  // clearly added footer
	inForm  bool
}

func NewUI(db *sqlx.DB) *UI {
	ui := &UI{
		app:     tview.NewApplication(),
		db:      db,
		list:    tview.NewList().ShowSecondaryText(true),
		details: tview.NewTextView(),
	}

	ui.details.
		SetDynamicColors(true).
		SetWrap(true).
		SetScrollable(true).
		SetRegions(true).
		SetChangedFunc(func() {
			ui.app.Draw()
		})

	ui.list.SetBorder(true).SetTitle(" Scripts ").SetBorderColor(tcell.ColorYellow)
	ui.details.SetBorder(true).SetTitle(" Details (↑/↓ Scroll) ").SetBorderColor(tcell.ColorWhite)

	ui.footer = tview.NewTextView()
	ui.footer.SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]Tab[white]: Switch Pane | [green]C[white]: Create Script | [red]D[white]: Delete Script | [blue]E[white]: Execute Script | [cyan]Ctrl+Q[white]: Quit")

	// Call SetBorder separately to avoid type mismatch
	ui.footer.SetBorder(true).SetBorderColor(tcell.ColorGray)

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
		ui.details.Clear()
		ui.details.
			SetDynamicColors(true).
			SetRegions(true).
			SetWrap(true).
			SetText(tview.TranslateANSI(highlightCode(script.Content, "bash")))
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
	ui.inForm = true

	var scriptContent string // temporarily store edited content here

	form := tview.NewForm()

	form.
		AddInputField("Name", "", 20, nil, nil).
		AddInputField("Description", "", 40, nil, nil).
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

			if scriptContent == "" {
				ui.inForm = false
				ui.details.SetText("[red]Script content cannot be empty. Please edit script content first.")
				return
			}

			script := database.Script{Name: name, Description: description, Content: scriptContent}
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

func (ui *UI) Run() error {
	ui.loadScripts()

	mainLayout := tview.NewFlex().
		AddItem(ui.list, 0, 1, true).
		AddItem(ui.details, 0, 2, false)

	// Explicit footer setup (fixed height: 1 or 3 lines, clearly visible)
	ui.footer.SetBorder(true).SetBorderColor(tcell.ColorGray)

	ui.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainLayout, 0, 1, true).
		AddItem(ui.footer, 3, 0, false) // explicitly give footer height=1, fixed height

	ui.app.SetFocus(ui.list)

	currentFocus := 0
	focusItems := []tview.Primitive{ui.list, ui.details}

	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ui.inForm {
			return event
		}

		if event.Key() == tcell.KeyCtrlQ {
			ui.app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyTab {
			currentFocus = (currentFocus + 1) % len(focusItems)
			ui.app.SetFocus(focusItems[currentFocus])

			if currentFocus == 0 {
				ui.list.SetBorderColor(tcell.ColorYellow)
				ui.details.SetBorderColor(tcell.ColorWhite)
			} else {
				ui.list.SetBorderColor(tcell.ColorWhite)
				ui.details.SetBorderColor(tcell.ColorYellow)
			}

			return nil
		}

		switch event.Rune() {
			case 'C', 'c':
				ui.showCreateForm()
			case 'D', 'd':
				ui.confirmDeleteScript()
			case 'E', 'e':
				ui.executeSelectedScript()
			}

			return event
		})

	return ui.app.SetRoot(ui.root, true).Run()
}




func (ui *UI) executeSelectedScript() {
	index := ui.list.GetCurrentItem()
	if index < 0 {
		return
	}

	scripts, err := database.GetScripts(ui.db)
	if err != nil {
		ui.details.SetText(fmt.Sprintf("[red]Failed to load scripts: %v", err))
		return
	}

	script := scripts[index]

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



func launchEditor(app *tview.Application, initialContent string) (string, error) {
	// Suspend the tview Application (restore terminal state)
	app.Suspend(func() {
		tmpfile, err := os.CreateTemp("", "bashhub-*.sh")
		if err != nil {
			return
		}
		defer os.Remove(tmpfile.Name())

		tmpfile.Write([]byte(initialContent))
		tmpfile.Close()

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano" // default fallback
		}

		cmd := exec.Command(editor, tmpfile.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return
		}

		updatedContent, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			return
		}

		initialContent = string(updatedContent)
	})

	return initialContent, nil
}

func highlightCode(code, language string) string {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal16m")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code // fallback to plain code
	}

	var buff bytes.Buffer
	err = formatter.Format(&buff, style, iterator)
	if err != nil {
		return code
	}

	return buff.String()
}
