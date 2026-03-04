
# AGENT.md

Guidelines for AI agents working on **ddev-xdebug-tui**

---

# Project Overview

This project implements a **minimal terminal debugger for PHP projects running in DDEV**.

The debugger connects to Xdebug using the **DBGp protocol** and provides a small terminal UI for:

- breakpoints
- stepping
- variable inspection
- stack inspection

This is intentionally **not a full Xdebug client**.

The goal is a **small, understandable tool** that can be launched quickly and understood by developers who prefer terminal workflows.

---

# Core Design Philosophy

This project prioritizes:

1. **Simplicity**
2. **Readability**
3. **Minimal dependencies**
4. **Small codebase**
5. **Predictable behavior**

The maintainer intends to read and understand every line of code before release.

Agents must **avoid introducing complexity** unless explicitly requested.

---

# Absolute Constraints

Do NOT introduce:

- conditional breakpoints
- watch expressions
- exception breakpoints
- multi-session handling
- concurrency unless absolutely necessary
- heavy frameworks
- complex architecture patterns

Do NOT attempt to implement the entire DBGp protocol.

Only the required subset should be implemented.

---

# Technology Stack

Language: Go  
UI Framework: tview  
Protocol: Minimal implementation of DBGp  
Distribution: DDEV add-on invoking a Go binary

---

# Supported DBGp Commands (PoC)

Only implement support for:

- breakpoint_set
- breakpoint_remove
- stack_get
- context_get
- step_into
- step_over
- step_out
- run

Do NOT implement additional commands unless necessary for debugging.

---

# Debugger Scope (PoC)

The debugger should support:

- one connection at a time
- ephemeral breakpoints
- line breakpoints only
- step in / step over / step out / continue
- viewing source
- viewing stack
- viewing variables for the current frame

The debugger should attach to **the first incoming Xdebug connection**.

Multiple simultaneous connections are intentionally not supported in the PoC.

---

# User Interface

The UI should be a **split-pane terminal layout**.

Suggested layout:

+------------------------------------------------+
| ddev-xdebug-tui | project: example             |
+-------------------+----------------------------+
| Stack             | Source                     |
|                   |                            |
|                   |                            |
+-------------------+----------------------------+
| Variables         | Breakpoints                |
+------------------------------------------------+
Command:

---

# Command Keys

n  step over  
s  step in  
o  step out  
c  continue  
q  quit  

Command input examples:

b file.php:45  
rb file.php:45  

---

# Architecture Rules

The project should be structured into simple modules:

cmd/
    main.go

internal/
    dbgclient/
    breakpoints/
    session/
    source/
    tui/

Responsibilities:

dbgclient
- TCP listener
- DBGp message handling

session
- manages current debug session

breakpoints
- stores breakpoints

source
- loads files from host filesystem

tui
- renders UI using tview

---

# Concurrency Policy

Prefer **single-threaded execution**.

If goroutines are used, they must be minimal and clearly documented.

Avoid complex synchronization.

---

# Code Quality Requirements

All code must be:

- small
- readable
- well-commented
- free of unnecessary abstraction

Prefer explicit code over clever code.

---

# Implementation Order

Agents should implement features in this order:

1. TCP listener
2. Accept Xdebug connection
3. Parse DBGp messages
4. Send DBGp commands
5. Basic stepping
6. Breakpoint handling
7. Source loading
8. Variable inspection
9. TUI layout
10. DDEV wrapper command

Each step should compile and run before moving to the next.

---

# Testing Strategy

Manual testing is acceptable for the PoC.

Typical test loop:

ddev debug
set breakpoint
visit page in browser
step through execution
inspect variables
quit debugger

---

# Things To Avoid

Agents should not:

- introduce large dependencies
- introduce unnecessary abstractions
- rewrite the architecture
- add new features not requested

If unsure, prefer the simplest possible implementation.

---

# Success Criteria

The project is successful if a developer can:

1. run `ddev debug`
2. set a breakpoint
3. trigger a request
4. step through code
5. inspect variables
6. quit the debugger

without using an IDE.

---

# Future Features (Not PoC)

These may be implemented later:

- conditional breakpoints
- multiple sessions
- watch expressions
- breakpoint persistence

Agents should NOT attempt these during PoC development.
