package dbgclient

import (
	"net"
)

// Listen starts a TCP listener on :9003 and accepts the first incoming
// connection. Additional connections are closed immediately (single-session
// policy). onConnect is called with the accepted connection.
// Listen is intended to be run in a goroutine.
func Listen(onConnect func(net.Conn)) error {
	listener, err := net.Listen("tcp", ":9003")
	if err != nil {
		return err
	}
	defer listener.Close()

	// Accept the first connection and invoke the callback.
	firstConn, err := listener.Accept()
	if err != nil {
		return err
	}
	onConnect(firstConn)

	// Close any subsequent connections immediately (single-session policy).
	// The listener stays open to accept and discard additional connections.
	for {
		conn, err := listener.Accept()
		if err != nil {
			// Listener closed or error occurred
			return err
		}
		conn.Close()
	}
}
