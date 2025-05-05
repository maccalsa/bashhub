package tui

import (
	"os"
	"os/exec"

	"github.com/rivo/tview"
)



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