package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/cellear/ddev-xdebug-tui/internal/dbgclient"
	"github.com/cellear/ddev-xdebug-tui/internal/tui"
)

func main() {
	app := tui.NewApp()

	// Start TCP listener in background goroutine.
	// When Xdebug connects, read the init packet, send step_into (break on entry),
	// then load and display the source file.
	go func() {
		err := dbgclient.Listen(func(conn net.Conn) {
			app.SetStatus("ddev-xdebug-tui | Xdebug connected")

			// Create a session — owns the persistent bufio.Reader for this connection.
			session := dbgclient.NewSession(conn)

			// Read the init packet Xdebug sends immediately on connect.
			data, err := session.ReadMessage()
			if err != nil {
				app.SetStatus("ddev-xdebug-tui | read error: " + err.Error())
				conn.Close()
				return
			}

			language, fileURI, err := dbgclient.ParseInit(data)
			if err != nil {
				app.SetStatus("ddev-xdebug-tui | parse error: " + err.Error())
				conn.Close()
				return
			}

			// Send step_into to pause at the first executable line (break on entry).
			_, err = session.StepInto()
			if err != nil {
				app.SetStatus("ddev-xdebug-tui | step error: " + err.Error())
				conn.Close()
				return
			}

			// Store the session so the input handler can access it
			app.SetSession(session)

			// Update status bar: "ddev-xdebug-tui | PHP | index.php | line 1"
			filename := session.CurrentFile
			if idx := strings.LastIndex(filename, "/"); idx >= 0 {
				filename = filename[idx+1:]
			}
			app.SetStatus(fmt.Sprintf("ddev-xdebug-tui | %s | %s | line %d",
				language, filename, session.CurrentLine))

			// Load and display source for the current paused location.
			app.refreshSource(session)

			// Keep connection open — session is ready for step commands (S3-3).
			_ = fileURI // will be used in future stories
		})
		if err != nil {
			app.SetStatus("ddev-xdebug-tui | listener error: " + err.Error())
		}
	}()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
