package tui

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/cellear/ddev-xdebug-tui/internal/breakpoints"
	"github.com/cellear/ddev-xdebug-tui/internal/dbgclient"
	"github.com/cellear/ddev-xdebug-tui/internal/source"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App represents the TUI application.
type App struct {
	app              *tview.Application
	statusBar        *tview.TextView
	sourcePanel      *tview.TextView
	stackPanel       *tview.TextView
	variablesPanel   *tview.TextView
	breakpointsPanel *tview.TextView
	mu               sync.Mutex
	session          *dbgclient.Session
	bpStore          breakpoints.Store
}

// NewApp creates and returns a new TUI application with the full split-pane layout.
func NewApp() *App {
	app := tview.NewApplication()

	grid := tview.NewGrid()
	grid.SetRows(1, 0, 0, 1)
	grid.SetColumns(25, 0)
	grid.SetBorders(true)

	statusBar := tview.NewTextView()
	statusBar.SetText("ddev-xdebug-tui v0.4 | waiting for Xdebug connection")
	statusBar.SetBackgroundColor(tcell.ColorBlue)
	statusBar.SetTextColor(tcell.ColorWhite)

	stackPanel := tview.NewTextView()
	stackPanel.SetBorder(true)
	stackPanel.SetTitle("Stack")
	stackPanel.SetText("")
	stackPanel.SetDynamicColors(true)

	sourcePanel := tview.NewTextView()
	sourcePanel.SetBorder(true)
	sourcePanel.SetTitle("Source")
	sourcePanel.SetText("")
	sourcePanel.SetDynamicColors(true)
	sourcePanel.SetRegions(true)
	sourcePanel.SetScrollable(true)
	sourcePanel.SetWrap(false)

	variablesPanel := tview.NewTextView()
	variablesPanel.SetBorder(true)
	variablesPanel.SetTitle("Variables")
	variablesPanel.SetText("")
	variablesPanel.SetDynamicColors(true)

	breakpointsPanel := tview.NewTextView()
	breakpointsPanel.SetBorder(true)
	breakpointsPanel.SetTitle("Breakpoints")
	breakpointsPanel.SetText("")

	topRow := tview.NewFlex()
	topRow.SetDirection(tview.FlexColumn)
	topRow.AddItem(stackPanel, 0, 1, false)
	topRow.AddItem(sourcePanel, 0, 3, false)

	bottomRow := tview.NewFlex()
	bottomRow.SetDirection(tview.FlexColumn)
	bottomRow.AddItem(variablesPanel, 0, 1, false)
	bottomRow.AddItem(breakpointsPanel, 0, 1, false)

	commandInput := tview.NewInputField()
	commandInput.SetLabel("Command: ")
	commandInput.SetLabelColor(tcell.ColorWhite)
	commandInput.SetFieldBackgroundColor(tcell.ColorBlack)
	commandInput.SetFieldTextColor(tcell.ColorWhite)

	grid.AddItem(statusBar, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(topRow, 1, 0, 1, 2, 0, 0, false)
	grid.AddItem(bottomRow, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(commandInput, 3, 0, 1, 2, 0, 0, true)

	app.SetRoot(grid, true)

	tuiApp := &App{
		app:              app,
		statusBar:        statusBar,
		sourcePanel:      sourcePanel,
		stackPanel:       stackPanel,
		variablesPanel:   variablesPanel,
		breakpointsPanel: breakpointsPanel,
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			app.Stop()
			return nil
		case 's', 'n', 'o', 'r':
			if commandInput.GetText() == "" {
				go tuiApp.handleCommand(string(event.Rune()))
				return nil
			}
		}
		return event
	})

	commandInput.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		cmd := strings.TrimSpace(commandInput.GetText())
		commandInput.SetText("")
		if cmd == "" {
			return
		}
		go tuiApp.handleCommand(cmd)
	})

	return tuiApp
}

// SetStatus updates the status bar text. Safe to call from any goroutine.
func (a *App) SetStatus(text string) {
	a.app.QueueUpdateDraw(func() {
		a.statusBar.SetText(text)
	})
}

// SetInitInfo displays the language and fileURI from Xdebug's init packet.
func (a *App) SetInitInfo(language, fileURI string) {
	a.app.QueueUpdateDraw(func() {
		a.sourcePanel.SetText(fmt.Sprintf("Language: %s\nFile:     %s", language, fileURI))
	})
}

// SetSource displays formatted source content in the Source panel.
func (a *App) SetSource(content string, currentLine int) {
	a.app.QueueUpdateDraw(func() {
		a.sourcePanel.SetText(content)
		if currentLine > 0 {
			a.sourcePanel.Highlight(fmt.Sprintf("%d", currentLine))
			a.sourcePanel.ScrollToHighlight()
		}
	})
}

// SetVariables updates the Variables panel. Safe to call from any goroutine.
func (a *App) SetVariables(text string) {
	a.app.QueueUpdateDraw(func() {
		a.variablesPanel.SetText(text)
	})
}

// SetStack updates the Stack panel. Safe to call from any goroutine.
func (a *App) SetStack(text string) {
	a.app.QueueUpdateDraw(func() {
		a.stackPanel.SetText(text)
	})
}

// SetBreakpoints updates the Breakpoints panel. Safe to call from any goroutine.
func (a *App) SetBreakpoints(text string) {
	a.app.QueueUpdateDraw(func() {
		a.breakpointsPanel.SetText(text)
	})
}

// SetSession stores the active debug session and immediately refreshes all panels.
func (a *App) SetSession(session *dbgclient.Session) {
	a.mu.Lock()
	a.session = session
	a.mu.Unlock()
	if session != nil {
		a.refreshAll(session)
	}
}

// ClearSession nils the session pointer. Panels are intentionally NOT cleared
// so the user can see the last debug state while waiting for the next connection.
func (a *App) ClearSession() {
	a.mu.Lock()
	a.session = nil
	a.mu.Unlock()
}

// getSession retrieves the stored debug session. Safe to call from any goroutine.
func (a *App) getSession() *dbgclient.Session {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.session
}

// parseFileAndLine parses "filename.php:N" into filename and line number.
func parseFileAndLine(arg string) (file string, line int, err error) {
	parts := strings.SplitN(arg, ":", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("expected file:line, got %q", arg)
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid line number %q", parts[1])
	}
	return parts[0], n, nil
}

// currentFileBase returns just the filename portion of the session's CurrentFile URI.
func currentFileBase(session *dbgclient.Session) string {
	f := session.CurrentFile
	if idx := strings.LastIndex(f, "/"); idx >= 0 {
		f = f[idx+1:]
	}
	return f
}

// handleCommand processes a command string entered by the user.
// Must be called in a goroutine (may block on network I/O).
func (a *App) handleCommand(cmd string) {
	session := a.getSession()
	if session == nil {
		a.SetStatus("ddev-xdebug-tui | not connected")
		return
	}

	var status string
	var err error

	switch {
	case cmd == "s":
		status, err = session.StepInto()
	case cmd == "n":
		status, err = session.StepOver()
	case cmd == "o":
		status, err = session.StepOut()
	case cmd == "r":
		status, err = session.Run()
	case strings.HasPrefix(cmd, "b "):
		// Set breakpoint: "b index.php:6" or shorthand "b 6" (uses current file)
		arg := strings.TrimPrefix(cmd, "b ")
		if _, numErr := strconv.Atoi(arg); numErr == nil {
			arg = currentFileBase(session) + ":" + arg
		}
		file, line, err := parseFileAndLine(arg)
		if err != nil {
			a.SetStatus("ddev-xdebug-tui | " + err.Error())
			return
		}
		containerURI := "file:///var/www/html/" + file
		id, err := session.SetBreakpoint(containerURI, line)
		if err != nil {
			a.SetStatus(fmt.Sprintf("ddev-xdebug-tui | breakpoint error: %s", err.Error()))
			return
		}
		a.bpStore.Add(file, line, id)
		a.SetBreakpoints(a.bpStore.Format())
		a.SetStatus(fmt.Sprintf("ddev-xdebug-tui | breakpoint set: %s:%d", file, line))
		return
	case strings.HasPrefix(cmd, "rb "):
		// Remove breakpoint: "rb index.php:6" or shorthand "rb 6" (uses current file)
		arg := strings.TrimPrefix(cmd, "rb ")
		if _, numErr := strconv.Atoi(arg); numErr == nil {
			arg = currentFileBase(session) + ":" + arg
		}
		file, line, err := parseFileAndLine(arg)
		if err != nil {
			a.SetStatus("ddev-xdebug-tui | " + err.Error())
			return
		}
		id, err := a.bpStore.Remove(file, line)
		if err != nil {
			a.SetStatus(fmt.Sprintf("ddev-xdebug-tui | %s", err.Error()))
			return
		}
		if err := session.RemoveBreakpoint(id); err != nil {
			a.SetStatus(fmt.Sprintf("ddev-xdebug-tui | remove error: %s", err.Error()))
			return
		}
		a.SetBreakpoints(a.bpStore.Format())
		a.SetStatus(fmt.Sprintf("ddev-xdebug-tui | breakpoint removed: %s:%d", file, line))
		return
	default:
		a.SetStatus(fmt.Sprintf("ddev-xdebug-tui | unknown command: %s", cmd))
		return
	}

	if err != nil {
		a.SetStatus(fmt.Sprintf("ddev-xdebug-tui | error: %s", err.Error()))
		return
	}

	if status == "stopping" || status == "stopped" {
		a.SetStatus("ddev-xdebug-tui | Script finished — waiting for next connection…")
		session.Close() // signal main.go to loop back to ln.Accept()
		return
	}

	// Update status bar with new position
	a.SetStatus(fmt.Sprintf("ddev-xdebug-tui | PHP | %s | line %d",
		currentFileBase(session), session.CurrentLine))

	// Refresh all panels
	a.refreshAll(session)
}

// refreshAll refreshes source, variables, and stack panels from the current session state.
func (a *App) refreshAll(session *dbgclient.Session) {
	a.refreshSource(session)
	a.refreshVariables(session)
	a.refreshStack(session)
}

// refreshSource maps the session's current container path to a host path,
// loads the source file, and updates the Source panel.
func (a *App) refreshSource(session *dbgclient.Session) {
	hostPath, err := source.MapPath(session.CurrentFile)
	if err != nil {
		a.SetSource(fmt.Sprintf("source not found: %s", err.Error()), 0)
		return
	}
	content, err := source.Format(hostPath, session.CurrentLine)
	if err != nil {
		a.SetSource(fmt.Sprintf("source error: %s", err.Error()), 0)
		return
	}
	a.SetSource(content, session.CurrentLine)
}

// refreshVariables fetches and displays local variables for the current frame.
func (a *App) refreshVariables(session *dbgclient.Session) {
	vars, err := session.ContextGet()
	if err != nil {
		a.SetVariables(fmt.Sprintf("error: %s", err.Error()))
		return
	}
	if len(vars) == 0 {
		a.SetVariables("(no variables)")
		return
	}
	var sb strings.Builder
	for _, v := range vars {
		fmt.Fprintf(&sb, "%s = %s\n", v.Name, v.Value)
	}
	a.SetVariables(sb.String())
}

// refreshStack fetches and displays the current call stack.
func (a *App) refreshStack(session *dbgclient.Session) {
	frames, err := session.StackGet()
	if err != nil {
		a.SetStack(fmt.Sprintf("error: %s", err.Error()))
		return
	}
	if len(frames) == 0 {
		a.SetStack("(empty)")
		return
	}
	var sb strings.Builder
	for _, f := range frames {
		filename := f.Filename
		if idx := strings.LastIndex(filename, "/"); idx >= 0 {
			filename = filename[idx+1:]
		}
		if f.Level == 0 {
			fmt.Fprintf(&sb, "► %s:%d\n", filename, f.Lineno)
		} else {
			fmt.Fprintf(&sb, "  %s:%d\n", filename, f.Lineno)
		}
	}
	a.SetStack(sb.String())
}

// Run starts the application event loop.
func (a *App) Run() error {
	return a.app.Run()
}
