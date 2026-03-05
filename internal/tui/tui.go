package tui

import (
	"fmt"
	"strings"
	"sync"

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
	breakpointsPanel *tview.TextView
	mu               sync.Mutex
	session          *dbgclient.Session
}

// NewApp creates and returns a new TUI application with the full split-pane layout.
func NewApp() *App {
	app := tview.NewApplication()

	// Create the main grid container
	grid := tview.NewGrid()
	grid.SetRows(1, 0, 0, 1)  // status bar, stack+source, variables+breakpoints, command input
	grid.SetColumns(25, 0)     // left panel (stack), right panel (source)
	grid.SetBorders(true)

	// Status bar at the top
	statusBar := tview.NewTextView()
	statusBar.SetText("ddev-xdebug-tui | waiting for Xdebug connection")
	statusBar.SetBackgroundColor(tcell.ColorBlue)
	statusBar.SetTextColor(tcell.ColorWhite)

	// Stack panel (top-left)
	stackPanel := tview.NewTextView()
	stackPanel.SetBorder(true)
	stackPanel.SetTitle("Stack")
	stackPanel.SetText("")

	// Source panel (top-right): dynamic colors and regions for line highlighting.
	sourcePanel := tview.NewTextView()
	sourcePanel.SetBorder(true)
	sourcePanel.SetTitle("Source")
	sourcePanel.SetText("")
	sourcePanel.SetDynamicColors(true)
	sourcePanel.SetRegions(true)
	sourcePanel.SetScrollable(true)

	// Variables panel (bottom-left)
	variablesPanel := tview.NewTextView()
	variablesPanel.SetBorder(true)
	variablesPanel.SetTitle("Variables")
	variablesPanel.SetText("")

	// Breakpoints panel (bottom-right)
	breakpointsPanel := tview.NewTextView()
	breakpointsPanel.SetBorder(true)
	breakpointsPanel.SetTitle("Breakpoints")
	breakpointsPanel.SetText("")

	// Top row: Stack | Source
	topRow := tview.NewFlex()
	topRow.SetDirection(tview.FlexColumn)
	topRow.AddItem(stackPanel, 0, 1, false)
	topRow.AddItem(sourcePanel, 0, 3, false)

	// Bottom row: Variables | Breakpoints
	bottomRow := tview.NewFlex()
	bottomRow.SetDirection(tview.FlexColumn)
	bottomRow.AddItem(variablesPanel, 0, 1, false)
	bottomRow.AddItem(breakpointsPanel, 0, 1, false)

	// Command input at the bottom
	commandInput := tview.NewInputField()
	commandInput.SetLabel("Command: ")
	commandInput.SetFieldBackgroundColor(tcell.ColorDefault)

	// Add items to grid
	grid.AddItem(statusBar, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(topRow, 1, 0, 1, 2, 0, 0, false)
	grid.AddItem(bottomRow, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(commandInput, 3, 0, 1, 2, 0, 0, true)

	// Set root and configure app
	app.SetRoot(grid, true)

	// Build the App struct first so closures below can reference it.
	tuiApp := &App{
		app:              app,
		statusBar:        statusBar,
		sourcePanel:      sourcePanel,
		breakpointsPanel: breakpointsPanel,
	}

	// Global key bindings.
	// Single-char step commands (s/n/o/r) fire immediately when the command
	// input is empty, so the user doesn't have to press Enter after each step.
	// When the input has text (e.g. typing "b index.php:10"), these keys are
	// passed through normally so they don't accidentally trigger a step.
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

	// Multi-word commands (e.g. breakpoints) are submitted via Enter in the input bar.
	commandInput.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		cmd := strings.TrimSpace(commandInput.GetText())
		commandInput.SetText("") // clear after submit
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
// Safe to call from any goroutine.
func (a *App) SetInitInfo(language, fileURI string) {
	a.app.QueueUpdateDraw(func() {
		a.sourcePanel.SetText(fmt.Sprintf("Language: %s\nFile:     %s", language, fileURI))
	})
}

// SetSource displays formatted source content in the Source panel and highlights
// the current line. content should be produced by source.Format().
// Safe to call from any goroutine.
func (a *App) SetSource(content string, currentLine int) {
	a.app.QueueUpdateDraw(func() {
		a.sourcePanel.SetText(content)
		if currentLine > 0 {
			a.sourcePanel.Highlight(fmt.Sprintf("%d", currentLine))
			a.sourcePanel.ScrollToHighlight()
		}
	})
}

// SetBreakpoints updates the Breakpoints panel with the given text.
// Safe to call from any goroutine.
func (a *App) SetBreakpoints(text string) {
	a.app.QueueUpdateDraw(func() {
		a.breakpointsPanel.SetText(text)
	})
}

// SetSession stores the active debug session and immediately refreshes the
// source panel to display the current paused location. Safe to call from any goroutine.
func (a *App) SetSession(session *dbgclient.Session) {
	a.mu.Lock()
	a.session = session
	a.mu.Unlock()
	if session != nil {
		a.refreshSource(session)
	}
}

// getSession retrieves the stored debug session. Safe to call from any goroutine.
func (a *App) getSession() *dbgclient.Session {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.session
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

	switch cmd {
	case "s":
		status, err = session.StepInto()
	case "n":
		status, err = session.StepOver()
	case "o":
		status, err = session.StepOut()
	case "r":
		status, err = session.Run()
	default:
		a.SetStatus(fmt.Sprintf("ddev-xdebug-tui | unknown command: %s", cmd))
		return
	}

	if err != nil {
		a.SetStatus(fmt.Sprintf("ddev-xdebug-tui | error: %s", err.Error()))
		return
	}

	if status == "stopping" || status == "stopped" {
		a.SetStatus("ddev-xdebug-tui | session ended")
		return
	}

	// Update status bar with new position
	filename := session.CurrentFile
	if idx := strings.LastIndex(filename, "/"); idx >= 0 {
		filename = filename[idx+1:]
	}
	a.SetStatus(fmt.Sprintf("ddev-xdebug-tui | PHP | %s | line %d", filename, session.CurrentLine))

	// Refresh source panel
	a.refreshSource(session)
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

// Run starts the application event loop.
func (a *App) Run() error {
	return a.app.Run()
}
