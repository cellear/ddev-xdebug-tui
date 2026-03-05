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

			// Read the init packet Xdebug sends immediately on connect.
			data, err := dbgclient.ReadMessage(conn)
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

			// Update status bar: "ddev-xdebug-tui | PHP | index.php"
			// Extract just the filename from the full URI path
			filename := fileURI
			if idx := strings.LastIndex(fileURI, "/"); idx >= 0 {
				filename = fileURI[idx+1:]
			}
			app.SetStatus(fmt.Sprintf("ddev-xdebug-tui | %s | %s", language, filename))
		})
		if err != nil {
			app.SetStatus("ddev-xdebug-tui | listener error: " + err.Error())
		}
	}()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
