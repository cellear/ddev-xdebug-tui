
# ddev-xdebug-tui_technical_architecture.md

## Overview
This document describes the internal architecture for the ddev-xdebug-tui debugger.

Core principles:
- Minimal DBGp protocol subset
- Single connection model
- Imperative control flow
- Small and understandable codebase

## Components
1. TCP Listener
2. DBGp Protocol Handler
3. Debug Session Manager
4. Breakpoint Manager
5. Source Renderer
6. Variable Inspector
7. TUI Layout (tview)

## Network Model
Xdebug connects from the DDEV container to the host machine on port 9003.

Flow:
PHP request → Xdebug → TCP → ddev-xdebug-tui → UI

## Session Lifecycle
1. Listener waits for connection
2. Xdebug connects
3. Session initialized
4. Breakpoints applied
5. Execution control begins
