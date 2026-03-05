package dbgclient

import (
	"bufio"
	"bytes"
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
	// Go's xml package only supports UTF-8. Xdebug declares iso-8859-1 but
	// the content is ASCII-compatible, so we rewrite the declaration before parsing.
	data = bytes.ReplaceAll(data, []byte(`encoding="iso-8859-1"`), []byte(`encoding="UTF-8"`))

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

// Session holds the state of a live Xdebug debug session.
type Session struct {
	conn        net.Conn
	reader      *bufio.Reader
	txID        int
	CurrentFile string // current paused file (container path, e.g. file:///var/www/html/index.php)
	CurrentLine int    // current paused line number
}

// NewSession creates a new Session wrapping the given connection.
func NewSession(conn net.Conn) *Session {
	return &Session{
		conn:   conn,
		reader: bufio.NewReader(conn),
		txID:   0,
	}
}

// SendCommand sends a DBGp command to Xdebug.
// Format: "<cmd> -i <txID>\0"
// txID is auto-incremented on each call.
func (s *Session) SendCommand(cmd string) error {
	s.txID++
	msg := fmt.Sprintf("%s -i %d\000", cmd, s.txID)
	_, err := s.conn.Write([]byte(msg))
	return err
}

// ReadMessage reads one complete DBGp message from the session connection.
// Uses the session's persistent bufio.Reader to avoid losing buffered bytes.
func (s *Session) ReadMessage() ([]byte, error) {
	// Read the length prefix: read bytes until we hit the first \0
	lengthStr := ""
	for {
		b, err := s.reader.ReadByte()
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
	n, err := io.ReadFull(s.reader, payload)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("EOF while reading payload (expected %d, got %d)", length, n)
		}
		return nil, fmt.Errorf("error reading payload: %w", err)
	}

	// Read and discard the final \0
	b, err := s.reader.ReadByte()
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

// stepResponse is the XML structure returned by step_into/step_over/step_out/run.
type stepResponse struct {
	XMLName xml.Name    `xml:"response"`
	Status  string      `xml:"status,attr"`
	Reason  string      `xml:"reason,attr"`
	Message stepMessage `xml:"https://xdebug.org/dbgp/xdebug message"`
}

// stepMessage represents the <xdebug:message> element in a step response.
type stepMessage struct {
	Filename string `xml:"filename,attr"`
	Lineno   int    `xml:"lineno,attr"`
}

// sendStep sends a step command and reads+parses the response.
// Updates s.CurrentFile and s.CurrentLine on success.
// Returns the response status ("break", "stopping", "stopped") and any error.
func (s *Session) sendStep(cmd string) (status string, err error) {
	if err := s.SendCommand(cmd); err != nil {
		return "", fmt.Errorf("send %s: %w", cmd, err)
	}
	data, err := s.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("read %s response: %w", cmd, err)
	}
	data = bytes.ReplaceAll(data, []byte(`encoding="iso-8859-1"`), []byte(`encoding="UTF-8"`))
	var resp stepResponse
	if err := xml.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("parse %s response: %w", cmd, err)
	}
	if resp.Message.Filename != "" {
		s.CurrentFile = resp.Message.Filename
		s.CurrentLine = resp.Message.Lineno
	}
	return resp.Status, nil
}

// StepInto sends step_into and updates session state.
func (s *Session) StepInto() (status string, err error) {
	return s.sendStep("step_into")
}
