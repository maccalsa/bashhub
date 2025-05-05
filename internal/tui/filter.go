package tui

import (
	"strings"

	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/maccalsa/bashhub/internal/database"
	"github.com/rivo/tview"
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

	// First group scripts clearly by category
	catMap := make(map[string][]database.Script)
	for _, script := range scripts {
		catMap[script.Category] = append(catMap[script.Category], script)
	}

	// Extract and sort categories explicitly (case insensitive)
	var categories []string
	for category := range catMap {
		categories = append(categories, category)
	}
	sort.Slice(categories, func(i, j int) bool {
		return strings.ToLower(categories[i]) < strings.ToLower(categories[j])
	})

	for _, category := range categories {
		categoryLower := strings.ToLower(category)
		var matchingScripts []database.Script

		// Check if category matches query
		if strings.Contains(categoryLower, query) {
			// clearly include all scripts in matching category
			matchingScripts = catMap[category]
		} else {
			// explicitly filter scripts matching query clearly
			for _, script := range catMap[category] {
				if strings.Contains(strings.ToLower(script.Name), query) ||
					strings.Contains(strings.ToLower(script.Description), query) {
					matchingScripts = append(matchingScripts, script)
				}
			}
		}

		if len(matchingScripts) > 0 {
			// explicitly sort scripts clearly within category
			sort.Slice(matchingScripts, func(i, j int) bool {
				return strings.ToLower(matchingScripts[i].Name) < strings.ToLower(matchingScripts[j].Name)
			})

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
