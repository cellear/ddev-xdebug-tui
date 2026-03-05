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
	// When Xdebug connects, read the init packet and display it.
	go func() {
		err := dbgclient.Listen(func(conn net.Conn) {
			app.SetStatus("ddev-xdebug-tui | Xdebug connected")

			// Create a session wrapping the connection
			session := dbgclient.NewSession(conn)

			// Read the init packet Xdebug sends immediately on connect.
			// Use session.ReadMessage() to use the persistent bufio.Reader
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

			app.SetInitInfo(language, fileURI)

			// Send step_into to pause at the first line (break on entry)
			status, err := session.StepInto()
			if err != nil {
				app.SetStatus("ddev-xdebug-tui | step error: " + err.Error())
				conn.Close()
				return
			}

			// Update status bar with file, line, and status
			filename := session.CurrentFile
			if idx := strings.LastIndex(filename, "/"); idx >= 0 {
				filename = filename[idx+1:]
			}
			app.SetStatus(fmt.Sprintf("ddev-xdebug-tui | %s | %s | line %d (status: %s)",
				language, filename, session.CurrentLine, status))
		})
		if err != nil {
			app.SetStatus("ddev-xdebug-tui | listener error: " + err.Error())
		}
	}()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
