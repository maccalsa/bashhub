package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/maccalsa/bashhub/internal/database"
	"github.com/rivo/tview"
	"fmt"
)

func (ui *UI) filterScripts(query string) {
	rootNode := tview.NewTreeNode(fmt.Sprintf("Search: '%s'", query)).SetColor(tcell.ColorYellow)
	ui.tree.SetRoot(rootNode).SetCurrentNode(rootNode)

	scripts, err := database.GetScripts(ui.db)
	if err != nil {
		ui.details.SetText(fmt.Sprintf("[red]Error loading scripts: %v", err))
		return
	}

	query = strings.ToLower(query)

	// First group scripts by category clearly
	catMap := make(map[string][]database.Script)
	for _, script := range scripts {
		catMap[script.Category] = append(catMap[script.Category], script)
	}

	for category, scripts := range catMap {
		categoryLower := strings.ToLower(category)

		var matchingScripts []database.Script

		// Check if category matches
		if strings.Contains(categoryLower, query) {
			// if category matches, clearly add ALL scripts in category
			matchingScripts = scripts
		} else {
			// else, clearly add only scripts matching query
			for _, script := range scripts {
				if strings.Contains(strings.ToLower(script.Name), query) ||
					strings.Contains(strings.ToLower(script.Description), query) {
					matchingScripts = append(matchingScripts, script)
				}
			}
		}

		if len(matchingScripts) > 0 {
			catNode := tview.NewTreeNode(category).SetColor(tcell.ColorGreen)
			for _, script := range matchingScripts {
				script := script // capture clearly
				scriptNode := tview.NewTreeNode(script.Name).
					SetReference(script).
					SetColor(tcell.ColorWhite)
				catNode.AddChild(scriptNode)
			}
			rootNode.AddChild(catNode)
		}
	}

	ui.tree.SetSelectedFunc(func(node *tview.TreeNode) {
		ref := node.GetReference()
		if ref != nil {
			script := ref.(database.Script)
			ui.details.SetText(highlightCode(script.Content, script.Language))
		} else {
			node.SetExpanded(!node.IsExpanded())
		}
	})
}