package tui

import (
	"fmt"
	"strings"

	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/maccalsa/bashhub/internal/database"
	"github.com/rivo/tview"
)

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

	// Extract categories and sort explicitly (case-insensitive)
	var categories []string
	for category := range catMap {
		categories = append(categories, category)
	}
	sort.Slice(categories, func(i, j int) bool {
		return strings.ToLower(categories[i]) < strings.ToLower(categories[j])
	})

	for _, category := range categories {
		// explicitly sort scripts within category
		sort.Slice(catMap[category], func(i, j int) bool {
			return strings.ToLower(catMap[category][i].Name) < strings.ToLower(catMap[category][j].Name)
		})

		catNode := tview.NewTreeNode(category).
			SetColor(tcell.ColorGreen)

		for _, script := range catMap[category] {
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
			ui.details.SetText(highlightCode(script.Content, script.Language))
		}
	})
}
