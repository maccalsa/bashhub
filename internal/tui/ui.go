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
	tree    *tview.TreeView
	details *tview.TextView
	root    tview.Primitive // root primitive
	footer  *tview.TextView  // clearly added footer
	inForm  bool
}

func NewUI(db *sqlx.DB) *UI {
	ui := &UI{
		app:     tview.NewApplication(),
		db:      db,
		tree:    tview.NewTreeView(),
		details: tview.NewTextView().SetDynamicColors(true),
	}

	ui.tree.SetBorder(true).SetTitle(" Scripts ")
	ui.details.SetBorder(true).SetTitle(" Details (↑/↓ Scroll) ")

	ui.footer = tview.NewTextView()
	ui.footer.SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]Tab[white]: Switch Pane | [green]C[white]: Create Script | [orange]X[white]: Edit Script | [red]D[white]: Delete Script | [blue]E[white]: Execute Script | [cyan]Ctrl+Q[white]: Quit")

	ui.footer.SetBorder(true).SetBorderColor(tcell.ColorGray)

	return ui
}

func (ui *UI) loadScripts() {
	rootNode := tview.NewTreeNode("Scripts").SetColor(tcell.ColorYellow)
	ui.tree.SetRoot(rootNode).SetCurrentNode(rootNode)

	scripts, err := database.GetScripts(ui.db)
	if err != nil {
		ui.details.SetText(fmt.Sprintf("[red]Error loading scripts: %v", err))
		return
	}

	catMap := make(map[string][]database.Script)
	for _, script := range scripts {
		catMap[script.Category] = append(catMap[script.Category], script)
	}

	for category, scripts := range catMap {
		catNode := tview.NewTreeNode(category).
			SetColor(tcell.ColorGreen)

		for _, script := range scripts {
			script := script // capture clearly
			scriptNode := tview.NewTreeNode(script.Name).
				SetReference(script).
				SetColor(tcell.ColorWhite).
				SetSelectable(true)
			catNode.AddChild(scriptNode)
		}
		rootNode.AddChild(catNode)
	}

	ui.tree.SetSelectedFunc(func(node *tview.TreeNode) {
		ref := node.GetReference()
		if ref == nil {
			node.SetExpanded(!node.IsExpanded())
		} else {
			script := ref.(database.Script)
			ui.details.SetText(highlightCode(script.Content, "bash"))
		}
	})
}

func (ui *UI) confirmDeleteScript() {
	node := ui.tree.GetCurrentNode()
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
			ui.app.SetRoot(ui.root, true)
		})

	ui.app.SetRoot(modal, false)
}


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

			script := database.Script{
				Name: name, 
				Description: description, 
				Content: scriptContent,
				Category: category,
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

func (ui *UI) Run() error {
	ui.loadScripts()

	mainLayout := tview.NewFlex().
		AddItem(ui.tree, 0, 1, true).
		AddItem(ui.details, 0, 2, false)

	ui.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainLayout, 0, 1, true).
		AddItem(ui.footer, 3, 0, false)

	ui.app.SetFocus(ui.tree)

	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ui.inForm {
			return event
		}

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

	// Explicitly switch to the robust terminal256 formatter for broader compatibility.
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	var buff bytes.Buffer
	err = formatter.Format(&buff, style, iterator)
	if err != nil {
		return code
	}

	return tview.TranslateANSI(buff.String())
}
