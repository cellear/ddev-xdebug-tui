package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/cellear/ddev-xdebug-tui/internal/dbgclient"
	"github.com/cellear/ddev-xdebug-tui/internal/source"
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

			// Update status bar: "ddev-xdebug-tui | PHP | index.php | line 1"
			filename := session.CurrentFile
			if idx := strings.LastIndex(filename, "/"); idx >= 0 {
				filename = filename[idx+1:]
			}
			app.SetStatus(fmt.Sprintf("ddev-xdebug-tui | %s | %s | line %d",
				language, filename, session.CurrentLine))

			// Load and display source for the current paused location.
			refreshSource(app, session)

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

// refreshSource maps the session's current container path to a host path,
// loads the source file, and updates the Source panel.
func refreshSource(app *tui.App, session *dbgclient.Session) {
	hostPath, err := source.MapPath(session.CurrentFile)
	if err != nil {
		app.SetSource(fmt.Sprintf("source not found: %s", err.Error()), 0)
		return
	}

	content, err := source.Format(hostPath, session.CurrentLine)
	if err != nil {
		app.SetSource(fmt.Sprintf("source error: %s", err.Error()), 0)
		return
	}

	app.SetSource(content, session.CurrentLine)
}
