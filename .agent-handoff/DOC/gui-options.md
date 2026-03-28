# GUI Options for ddev-xdebug-tui

This document captures an analysis of possible UI directions for the project — conducted 2026-03-28 — as a reference for future decision-making. No changes were made to the codebase. These are options to evaluate.

---

## Current UI

The project uses **tview** (built on **tcell**) for a split-pane terminal UI. The layout is:

- Status bar (blue/white)
- Stack pane (top-left)
- Source pane (top-right, scrollable)
- Variables pane (bottom-left)
- Breakpoints pane (bottom-right)
- Command input (bottom)

Key files: `internal/tui/tui.go`

---

## Option 1: Colored Panes (Current TUI, Cosmetic Change)

**Effort: ~15 minutes. No architectural changes.**

tview/tcell already supports per-widget background and text colors. The status bar already uses this (`tcell.ColorBlue`). Any pane can be styled with:

```go
stackPanel.SetBackgroundColor(tcell.ColorDarkSlateGray)
sourcePanel.SetBackgroundColor(tcell.ColorDarkBlue)
variablesPanel.SetBackgroundColor(tcell.ColorDarkGreen)
breakpointsPanel.SetBackgroundColor(tcell.ColorDarkRed)
```

Color options available:

- **Named colors** — `tcell.ColorNavy`, `tcell.ColorDarkCyan`, etc.
- **True-color RGB** — `tcell.NewRGBColor(r, g, b)` — any RGB value; requires a true-color terminal (most modern terminals qualify)
- **256-color palette** — `tcell.Color256` for broader compatibility

Border and title text can also be styled independently with `SetBorderColor()` and `SetTitleColor()`.

Since several panels already have `.SetDynamicColors(true)`, inline tview color tags (e.g. `[yellow]`) can be embedded in text content for syntax highlighting.

**Verdict:** Easy win. Keeps the tool's minimal-dependency philosophy intact.

---

## Option 2: Desktop GUI (Native Window with Mouse Support)

These replace the terminal entirely with a real windowed application. All options below are pure Go.

| Library | Style | Pros | Cons |
|---------|-------|------|------|
| **[Fyne](https://fyne.io)** | Material-ish cross-platform | Most popular Go GUI lib; active community; good widget set; resizable panes; mouse support out of the box | Adds ~20 MB to binary; custom rendering (not truly native) |
| **[Gio](https://gioui.org)** | Immediate-mode GPU-rendered | Very performant; modern architecture | Steeper learning curve; fewer pre-built widgets |
| **[Wails](https://wails.io)** | Go backend + web frontend in native window | Small binary; use React/Vue/Svelte for UI; single distributable | Two-language stack; requires web-dev knowledge |
| **[gotk4](https://github.com/diamondburned/gotk4)** | GTK4 native bindings | Truly native look on Linux | GTK dependency; macOS support is awkward |
| **[walk](https://github.com/lxn/walk)** | Win32 native bindings | Native Windows look | Windows-only |

### Recommended: Fyne

For a debugger with panes, a source viewer, variable inspector, and command input, **Fyne** is the most practical migration path:

- `container.NewHSplit` / `VSplit` for the pane layout
- `widget.List` for stack and variables
- `widget.Entry` for the command input
- Resizable panes and mouse interactions out of the box
- Cross-platform: macOS, Windows, Linux

### Architectural impact

The backend (`internal/dbgclient`, `internal/breakpoints`, `internal/source`) would be unchanged. The `internal/tui` package would be replaced with `internal/gui`. However, the current `tui.go` file mixes rendering, command handling, and debug-session refresh logic — before migrating to any new UI framework, it would be worth extracting a small controller/service layer that both the TUI and a future GUI could call.

---

## Option 3: Web/HTML Interface (React or Plain HTML)

### Option 3a: Wails (native window, web frontend)

[Wails v2](https://wails.io) wraps a web frontend in a native OS window. No browser required; the output is a single binary.

- Go backend exposes functions that the JavaScript frontend calls directly (type-safe bindings generated automatically)
- Frontend can be React, Vue, Svelte, or plain HTML/CSS/JS
- Ideal for: rich syntax highlighting (CodeMirror or Monaco), responsive pane layouts, familiar web-dev tooling during development
- Used by several developer tools as an Electron alternative

### Option 3b: Embedded HTTP server + browser tab

The Go app starts a local HTTP server and optionally opens the user's default browser:

- Frontend is React (or plain HTML) served as static assets embedded in the binary via `go:embed`
- Real-time debug events stream to the browser via **WebSocket** (natural fit for step/breakpoint/variable updates)
- Pros: full power of web dev tooling; CodeMirror/Monaco for syntax highlighting; browser DevTools for debugging the UI itself
- Cons: less integrated feel — it's a browser tab, not a window; port conflict edge cases

This pattern is used by tools like Delve's headless mode, Clockwork (PHP), and many Go-based devtools.

### Recommended: Wails (if going the web route)

Wails gives the polished "desktop app" feel while letting you write the frontend in React. The Go backend (`dbgclient` etc.) stays essentially unchanged.

---

## Summary

| Option | Effort | Dependency impact | Architectural change |
|--------|--------|-------------------|----------------------|
| Colored TUI panes | ~15 min | None | None |
| Fyne desktop GUI | Days–weeks | Adds Fyne (~20 MB) | Replace `internal/tui` |
| Wails + React | Days–weeks | Adds Wails + JS toolchain | Replace `internal/tui`; add JS frontend |
| Go HTTP + WebSocket + React | Days–weeks | Adds JS toolchain | Add HTTP server; replace `internal/tui` |

The colored panes option is fully consistent with the project's minimal-dependency philosophy. The GUI/web options are meaningful scope expansions and would require re-examining the architecture constraints in `AGENT.md` before proceeding.

---

*Last updated: 2026-03-28 by cursor — analysis by claude-opus-4-6, documented by claude-sonnet-4-6*
