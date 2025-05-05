package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/maccalsa/bashhub/internal/database"
	"github.com/maccalsa/bashhub/internal/executor"
	"github.com/spf13/cobra"
)

var placeholderInputs []string

var runCmd = &cobra.Command{
	Use:   "run [script-name]",
	Short: "Run a bash script by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scriptName := args[0]
		db := database.ConnectDB()

		scripts, err := database.GetScripts(db)
		if err != nil {
			log.Fatalf("Failed to load scripts: %v", err)
		}

		var selectedScript *database.Script
		for _, script := range scripts {
			if script.Name == scriptName {
				selectedScript = &script
				break
			}
		}

		if selectedScript == nil {
			log.Fatalf("Script '%s' not found", scriptName)
		}

		// Parse placeholder values provided via --set flags
		inputs := parsePlaceholderInputs(placeholderInputs)

		placeholders := executor.ParsePlaceholders(selectedScript.Content)

		reader := bufio.NewReader(os.Stdin)

		for _, ph := range placeholders {
			if _, ok := inputs[ph]; !ok {
				fmt.Printf("%s: ", ph)
				text, _ := reader.ReadString('\n')
				inputs[ph] = strings.TrimSpace(text)
			}
		}

		finalScript := executor.ReplacePlaceholders(selectedScript.Content, inputs)

		err = executor.ExecuteScript(finalScript, func(line string) {
			fmt.Println(line)
		})

		if err != nil {
			log.Fatalf("Script execution failed: %v", err)
		}
	},
}

func parsePlaceholderInputs(inputPairs []string) map[string]string {
	inputs := make(map[string]string)
	for _, pair := range inputPairs {
		split := strings.SplitN(pair, "=", 2)
		if len(split) == 2 {
			inputs[split[0]] = split[1]
		}
	}
	return inputs
}

func init() {
	runCmd.Flags().StringArrayVarP(&placeholderInputs, "set", "s", []string{}, "Set placeholder values (key=value)")
	rootCmd.AddCommand(runCmd)
}
