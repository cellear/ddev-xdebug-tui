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
	// dbgclient.Listen now loops: after each session ends (onConnect returns),
	// it accepts the next connection automatically.
	go func() {
		err := dbgclient.Listen(func(conn net.Conn) {
			app.SetStatus("ddev-xdebug-tui | Xdebug connected")

			session := dbgclient.NewSession(conn)

			data, err := session.ReadMessage()
			if err != nil {
				app.SetStatus("ddev-xdebug-tui | read error: " + err.Error())
				conn.Close()
				return
			}

			language, _, err := dbgclient.ParseInit(data)
			if err != nil {
				app.SetStatus("ddev-xdebug-tui | parse error: " + err.Error())
				conn.Close()
				return
			}

			_, err = session.StepInto()
			if err != nil {
				app.SetStatus("ddev-xdebug-tui | step error: " + err.Error())
				conn.Close()
				return
			}

			// Store session and refresh all panels (source, variables, stack).
			app.SetSession(session)

			filename := session.CurrentFile
			if idx := strings.LastIndex(filename, "/"); idx >= 0 {
				filename = filename[idx+1:]
			}
			app.SetStatus(fmt.Sprintf("ddev-xdebug-tui | %s | %s | line %d",
				language, filename, session.CurrentLine))

			// Block until handleCommand calls session.Close() on stopping/stopped.
			// This keeps onConnect alive so Listen doesn't loop prematurely.
			<-session.Done

			// Nil the session; panels intentionally left showing last state.
			app.ClearSession()
		})
		if err != nil {
			app.SetStatus("ddev-xdebug-tui | listener error: " + err.Error())
		}
	}()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
