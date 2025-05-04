package executor

import (
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

func ExecuteScript(scriptContent string, outputHandler func(string)) error {
	cmd := exec.Command("bash", "-c", scriptContent)

	// Start the command with a pty. (pseudo terminal)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer ptmx.Close()

	// Copy PTY output to outputHandler
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				outputHandler(string(buf[:n]))
			}
			if err != nil {
				if err != io.EOF {
					outputHandler(err.Error())
				}
				break
			}
		}
	}()

	// Copy user input from terminal to the command
	go func() {
		io.Copy(ptmx, os.Stdin)
	}()

	return cmd.Wait()
}

