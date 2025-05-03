package main

import (
	"log"

	"github.com/maccalsa/bashhub/internal/database"
	"github.com/maccalsa/bashhub/internal/tui"
)

func main() {
	db := database.ConnectDB()
	ui := tui.NewUI(db)

	if err := ui.Run(); err != nil {
		log.Fatalf("Failed to run UI: %v", err)
	}
}
