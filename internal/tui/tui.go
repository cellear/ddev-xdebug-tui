package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App represents the TUI application.
type App struct {
	app       *tview.Application
	statusBar *tview.TextView
}

// NewApp creates and returns a new TUI application with the full split-pane layout.
func NewApp() *App {
	app := tview.NewApplication()
	
	// Create the main grid container
	grid := tview.NewGrid()
	grid.SetRows(1, 0, 0, 1)           // status bar, stack+source, variables+breakpoints, command input
	grid.SetColumns(25, 0)              // left panel (stack), right panel (source)
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
	
	// Source panel (top-right)
	sourcePanel := tview.NewTextView()
	sourcePanel.SetBorder(true)
	sourcePanel.SetTitle("Source")
	sourcePanel.SetText("")
	
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
	
	// Handle key bindings
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			app.Stop()
			return nil
		}
		return event
	})
	
	return &App{
		app:       app,
		statusBar: statusBar,
	}
}

// SetStatus updates the status bar text. Safe to call from any goroutine.
// Uses QueueUpdateDraw to avoid race conditions when called from background threads.
func (a *App) SetStatus(text string) {
	a.app.QueueUpdateDraw(func() {
		a.statusBar.SetText(text)
	})
}

// Run starts the application event loop.
func (a *App) Run() error {
	return a.app.Run()
}
