package main

import (
	"net"

	"github.com/cellear/ddev-xdebug-tui/internal/dbgclient"
	"github.com/cellear/ddev-xdebug-tui/internal/tui"
)

func main() {
	app := tui.NewApp()

	// Start TCP listener in background goroutine.
	// When Xdebug connects, update the status bar to show the connection.
	go func() {
		err := dbgclient.Listen(func(conn net.Conn) {
			defer conn.Close()
			app.SetStatus("ddev-xdebug-tui | Xdebug connected")
		})
		if err != nil {
			app.SetStatus("ddev-xdebug-tui | listener error: " + err.Error())
		}
	}()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
