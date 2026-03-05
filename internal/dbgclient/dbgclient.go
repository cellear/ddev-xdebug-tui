package dbgclient

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"strconv"
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

// ReadMessage reads one complete DBGp message from conn.
// DBGp framing: <decimal-length>\0<xml-payload>\0
// Returns the XML payload bytes, or an error.
func ReadMessage(conn net.Conn) ([]byte, error) {
	reader := bufio.NewReader(conn)

	// Read the length prefix: read bytes until we hit the first \0
	lengthStr := ""
	for {
		b, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("EOF while reading length prefix")
			}
			return nil, fmt.Errorf("error reading length prefix: %w", err)
		}
		if b == 0 {
			// Found the null terminator
			break
		}
		lengthStr += string(b)
	}

	// Parse the length string as an integer
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid length prefix: %q", lengthStr)
	}

	// Read exactly `length` bytes for the XML payload
	payload := make([]byte, length)
	n, err := io.ReadFull(reader, payload)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("EOF while reading payload (expected %d, got %d)", length, n)
		}
		return nil, fmt.Errorf("error reading payload: %w", err)
	}

	// Read and discard the final \0
	b, err := reader.ReadByte()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("EOF while reading final null terminator")
		}
		return nil, fmt.Errorf("error reading final null terminator: %w", err)
	}
	if b != 0 {
		return nil, fmt.Errorf("expected null terminator, got: %c", b)
	}

	return payload, nil
}

// initPacket represents the structure of an Xdebug <init> XML packet.
type initPacket struct {
	XMLName  xml.Name `xml:"init"`
	Language string   `xml:"language,attr"`
	FileURI  string   `xml:"fileuri,attr"`
}

// ParseInit parses an Xdebug <init> XML packet.
// Returns language and fileuri attributes, or an error.
func ParseInit(data []byte) (language string, fileURI string, err error) {
	var packet initPacket
	err = xml.Unmarshal(data, &packet)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse init packet: %w", err)
	}

	if packet.Language == "" {
		return "", "", fmt.Errorf("language attribute not found in init packet")
	}
	if packet.FileURI == "" {
		return "", "", fmt.Errorf("fileuri attribute not found in init packet")
	}

	return packet.Language, packet.FileURI, nil
}
