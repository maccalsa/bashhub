package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/maccalsa/bashhub/internal/tui"
	"github.com/maccalsa/bashhub/internal/database"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"log"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import scripts from a folder",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db := database.ConnectDB()
		importFolder(db, args[0])
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}

func importFolder(db *sqlx.DB, folderPath string) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		log.Fatalf("Failed to read folder: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue // Ignore directories explicitly
		}

		filePath := filepath.Join(folderPath, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		
		scriptName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		scriptLanguage := tui.DetectLanguage(string(content))

		script := database.Script{
			Name: scriptName,
			Description: fmt.Sprintf("Imported from %s", file.Name()),
			Content:     string(content),
			Category:    "Imported",
			Language:    scriptLanguage,
		}

		if err := database.CreateScript(db, script); err != nil {
			log.Printf("Failed to import script %s: %v", scriptName, err)
		} else {
			log.Printf("Successfully imported script %s", scriptName)
		}
	}
}


