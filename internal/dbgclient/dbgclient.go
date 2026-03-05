package dbgclient

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

// Listen starts a TCP listener on :9003 and calls onConnect for each incoming
// connection sequentially. onConnect should block until the session ends.
// Listen is intended to be run in a goroutine.
func Listen(onConnect func(net.Conn)) error {
	listener, err := net.Listen("tcp", ":9003")
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		onConnect(conn) // blocks until session ends; then loops to accept next
	}
}

// ReadMessage reads one complete DBGp message from conn.
// DBGp framing: <decimal-length>\0<xml-payload>\0
// Returns the XML payload bytes, or an error.
func ReadMessage(conn net.Conn) ([]byte, error) {
	reader := bufio.NewReader(conn)

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
			break
		}
		lengthStr += string(b)
	}

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid length prefix: %q", lengthStr)
	}

	payload := make([]byte, length)
	n, err := io.ReadFull(reader, payload)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("EOF while reading payload (expected %d, got %d)", length, n)
		}
		return nil, fmt.Errorf("error reading payload: %w", err)
	}

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
func ParseInit(data []byte) (language string, fileURI string, err error) {
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
	CurrentFile string // current paused file (container URI, e.g. file:///var/www/html/index.php)
	CurrentLine int    // current paused line number
	Done        chan struct{} // closed when session ends (status=stopping/stopped)
	closeOnce   sync.Once
}

// NewSession creates a new Session wrapping the given connection.
func NewSession(conn net.Conn) *Session {
	return &Session{
		conn:   conn,
		reader: bufio.NewReader(conn),
		txID:   0,
		Done:   make(chan struct{}),
	}
}

// Close signals that this session has ended. Safe to call multiple times.
func (s *Session) Close() {
	s.closeOnce.Do(func() {
		close(s.Done)
	})
}

// SendCommand sends a DBGp command to Xdebug.
// Format: "<cmd> -i <txID>\0"
func (s *Session) SendCommand(cmd string) error {
	s.txID++
	msg := fmt.Sprintf("%s -i %d\000", cmd, s.txID)
	_, err := s.conn.Write([]byte(msg))
	return err
}

// ReadMessage reads one complete DBGp message from the session connection.
func (s *Session) ReadMessage() ([]byte, error) {
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
			break
		}
		lengthStr += string(b)
	}

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid length prefix: %q", lengthStr)
	}

	payload := make([]byte, length)
	n, err := io.ReadFull(s.reader, payload)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("EOF while reading payload (expected %d, got %d)", length, n)
		}
		return nil, fmt.Errorf("error reading payload: %w", err)
	}

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

type stepMessage struct {
	Filename string `xml:"filename,attr"`
	Lineno   int    `xml:"lineno,attr"`
}

type breakpointSetResponse struct {
	XMLName xml.Name `xml:"response"`
	ID      string   `xml:"id,attr"`
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

func (s *Session) StepInto() (status string, err error) { return s.sendStep("step_into") }
func (s *Session) StepOver() (status string, err error) { return s.sendStep("step_over") }
func (s *Session) StepOut() (status string, err error)  { return s.sendStep("step_out") }
func (s *Session) Run() (status string, err error)      { return s.sendStep("run") }

// SetBreakpoint sends breakpoint_set for a line breakpoint.
func (s *Session) SetBreakpoint(fileURI string, line int) (id string, err error) {
	cmd := fmt.Sprintf("breakpoint_set -t line -f %s -n %d", fileURI, line)
	if err := s.SendCommand(cmd); err != nil {
		return "", fmt.Errorf("send breakpoint_set: %w", err)
	}
	data, err := s.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("read breakpoint_set response: %w", err)
	}
	data = bytes.ReplaceAll(data, []byte(`encoding="iso-8859-1"`), []byte(`encoding="UTF-8"`))
	var resp breakpointSetResponse
	if err := xml.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("parse breakpoint_set response: %w", err)
	}
	if resp.ID == "" {
		return "", fmt.Errorf("breakpoint_set response missing id attribute")
	}
	return resp.ID, nil
}

// RemoveBreakpoint sends breakpoint_remove for the given Xdebug breakpoint ID.
func (s *Session) RemoveBreakpoint(id string) error {
	cmd := fmt.Sprintf("breakpoint_remove -d %s", id)
	if err := s.SendCommand(cmd); err != nil {
		return fmt.Errorf("send breakpoint_remove: %w", err)
	}
	_, err := s.ReadMessage()
	return err
}

// Variable represents a single variable from context_get.
type Variable struct {
	Name  string
	Type  string
	Value string // formatted display value
}

// contextGetResponse is the XML structure returned by context_get.
type contextGetResponse struct {
	XMLName    xml.Name          `xml:"response"`
	Properties []contextProperty `xml:"property"`
}

type contextProperty struct {
	Name        string `xml:"name,attr"`
	Type        string `xml:"type,attr"`
	ClassName   string `xml:"classname,attr"`
	NumChildren int    `xml:"numchildren,attr"`
	Value       string `xml:",chardata"`
}

// ContextGet fetches local variables for the current stack frame (depth 0).
func (s *Session) ContextGet() ([]Variable, error) {
	if err := s.SendCommand("context_get -d 0"); err != nil {
		return nil, fmt.Errorf("send context_get: %w", err)
	}
	data, err := s.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read context_get response: %w", err)
	}
	data = bytes.ReplaceAll(data, []byte(`encoding="iso-8859-1"`), []byte(`encoding="UTF-8"`))
	var resp contextGetResponse
	if err := xml.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse context_get response: %w", err)
	}

	vars := make([]Variable, 0, len(resp.Properties))
	for _, p := range resp.Properties {
		v := Variable{Name: p.Name, Type: p.Type}
		switch p.Type {
		case "object":
			v.Value = fmt.Sprintf("{%s} (%d props)", p.ClassName, p.NumChildren)
		case "array":
			v.Value = fmt.Sprintf("[%d]", p.NumChildren)
		case "null":
			v.Value = "null"
		default:
			v.Value = strings.TrimSpace(p.Value)
		}
		vars = append(vars, v)
	}
	return vars, nil
}

// Frame represents a single stack frame from stack_get.
type Frame struct {
	Level    int
	Filename string // container URI (e.g. file:///var/www/html/index.php)
	Lineno   int
	Where    string
}

// stackGetResponse is the XML structure returned by stack_get.
type stackGetResponse struct {
	XMLName xml.Name     `xml:"response"`
	Frames  []stackFrame `xml:"stack"`
}

type stackFrame struct {
	Level    int    `xml:"level,attr"`
	Filename string `xml:"filename,attr"`
	Lineno   int    `xml:"lineno,attr"`
	Where    string `xml:"where,attr"`
}

// StackGet fetches the current call stack.
func (s *Session) StackGet() ([]Frame, error) {
	if err := s.SendCommand("stack_get"); err != nil {
		return nil, fmt.Errorf("send stack_get: %w", err)
	}
	data, err := s.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read stack_get response: %w", err)
	}
	data = bytes.ReplaceAll(data, []byte(`encoding="iso-8859-1"`), []byte(`encoding="UTF-8"`))
	var resp stackGetResponse
	if err := xml.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse stack_get response: %w", err)
	}

	frames := make([]Frame, len(resp.Frames))
	for i, f := range resp.Frames {
		frames[i] = Frame{
			Level:    f.Level,
			Filename: f.Filename,
			Lineno:   f.Lineno,
			Where:    f.Where,
		}
	}
	return frames, nil
}
