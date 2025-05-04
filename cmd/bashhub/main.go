package main

import (
	"log"

	"github.com/maccalsa/bashhub/internal/database"
	"github.com/maccalsa/bashhub/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bashhub",
	Short: "BashHub is a dynamic script execution manager",
	Run: func(cmd *cobra.Command, args []string) {
		// This is the default command (launches the TUI)
		db := database.ConnectDB()
		ui := tui.NewUI(db)
		if err := ui.Run(); err != nil {
			log.Fatalf("Failed to run UI: %v", err)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
