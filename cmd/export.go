package cmd

import (
	"github.com/maccalsa/bashhub/internal/database"
	"github.com/spf13/cobra"
	"log"
	"os"
	"bufio"
	"fmt"
	"strings"
	"github.com/jmoiron/sqlx"
	"github.com/maccalsa/bashhub/internal/executor"
)

var placeholderOverrides []string

var exportCmd = &cobra.Command{
	Use:   "export [script-name]",
	Short: "Export a script with placeholder substitutions",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db := database.ConnectDB()
		exportScript(db, args[0], placeholderOverrides)
	},
}

func init() {
	exportCmd.Flags().StringArrayVarP(&placeholderOverrides, "set", "s", []string{}, "Set placeholder values (key=value)")
	rootCmd.AddCommand(exportCmd)
}

func exportScript(db *sqlx.DB, scriptName string, overrides []string) {
	script, err := database.GetScriptByName(db, scriptName)
	if err != nil {
		log.Fatalf("Script not found: %v", err)
	}

	placeholders := executor.ParsePlaceholders(script.Content)
	inputs := make(map[string]string)

	// First, parse command-line overrides
	for _, override := range overrides {
		parts := strings.SplitN(override, "=", 2)
		if len(parts) != 2 {
			log.Fatalf("Invalid placeholder override: %s. Use key=value format.", override)
		}
		inputs[parts[0]] = parts[1]
	}

	reader := bufio.NewReader(os.Stdin)

	for _, ph := range placeholders {
		if _, exists := inputs[ph]; !exists {
			fmt.Printf("Enter value for '%s': ", ph)
			input, _ := reader.ReadString('\n')
			inputs[ph] = strings.TrimSpace(input)
		}
	}

	finalContent := executor.ReplacePlaceholders(script.Content, inputs)

	fmt.Println(finalContent)
}
