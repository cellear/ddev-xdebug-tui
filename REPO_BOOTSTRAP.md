
# REPO_BOOTSTRAP.md

Instructions for AI agents to bootstrap the **ddev-xdebug-tui** repository.

This file describes how to create the initial working skeleton of the project.
The goal is to reach a **compilable, runnable program as early as possible**.

Do NOT implement the full debugger immediately.
Build the system incrementally.

---

# Step 1 — Initialize Repository

Create the following structure:

ddev-xdebug-tui/
    go.mod
    AGENT.md
    REPO_BOOTSTRAP.md
    README.md

    cmd/
        ddev-xdebug-tui/
            main.go

    internal/
        dbgclient/
        session/
        breakpoints/
        source/
        tui/

The project should compile after Step 2.

---

# Step 2 — Initialize Go Module

Run:

go mod init github.com/<username>/ddev-xdebug-tui

Add dependency:

github.com/rivo/tview

Also add:

github.com/gdamore/tcell/v2

These are required for the TUI.

---

# Step 3 — Minimal Runnable Program

Create cmd/ddev-xdebug-tui/main.go

The program should:

1. Start a basic tview application
2. Render a simple box containing the text:

    ddev-xdebug-tui
    waiting for Xdebug connection

Example behavior:

Running:

go run ./cmd/ddev-xdebug-tui

Should display a full-screen terminal UI.

Do NOT implement debugger logic yet.

---

# Step 4 — TCP Listener

Implement a minimal TCP listener inside:

internal/dbgclient/

The listener should:

- listen on port 9003
- wait for an incoming connection
- print a message in the UI when a client connects

This confirms Xdebug can connect.

Example flow:

start program
visit a PHP page with Xdebug enabled
Xdebug connects
UI updates with:

    "Xdebug client connected"

---

# Step 5 — DBGp Message Capture

Once a connection exists:

- read raw DBGp messages from the socket
- print them to a debug log panel in the UI

Do NOT parse them yet.

Goal: verify message flow.

---

# Step 6 — Minimal DBGp Parser

Implement a simple parser that extracts:

- command
- transaction id
- XML payload

Only parse enough to understand:

init
stack_get
context_get

---

# Step 7 — Basic Debug Session

Create:

internal/session/

This module should:

- store the active connection
- track current file and line
- store stack frames

Keep this simple.

---

# Step 8 — Breakpoint Manager

Create:

internal/breakpoints/

Store breakpoints in memory:

map[string][]int

Where:

key = filename
value = list of line numbers

Breakpoints are ephemeral.

---

# Step 9 — Source Loader

Create:

internal/source/

This module should:

- load files from the host filesystem
- return surrounding lines for display

For example:

GetContext(file, line, radius)

Returns ~10 lines around the execution point.

---

# Step 10 — TUI Layout

Create:

internal/tui/

Layout should include:

Left panel:
Stack

Right panel:
Source

Bottom left:
Variables

Bottom right:
Breakpoints

Bottom line:
Command input

Use tview primitives:

Grid
TextView
InputField

---

# Step 11 — Step Commands

Implement commands:

n  step_over
s  step_into
o  step_out
c  run

Send these commands through the DBGp connection.

---

# Step 12 — Breakpoint Commands

User input:

b file.php:45

Agent should send:

breakpoint_set

Removing:

rb file.php:45

Send:

breakpoint_remove

---

# Step 13 — Manual Test Loop

Typical test process:

ddev xdebug on
ddev debug

visit site in browser

Debugger should:

pause at breakpoint
allow stepping
display source and variables

---

# Implementation Principles

Agents must:

- keep code small
- keep modules simple
- avoid unnecessary abstractions
- avoid concurrency unless required

The maintainer must be able to understand the entire codebase.

---

# End Goal

A developer should be able to:

run `ddev debug`
set a breakpoint
step through PHP code
inspect variables
exit debugger

without using an IDE.
