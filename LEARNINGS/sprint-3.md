# Sprint 3 Learnings: Go for PHP Developers

This document explains the Go concepts introduced in Sprint 3. If you haven't
read `LEARNINGS/sprint-1.md` and `LEARNINGS/sprint-2.md` yet, start there.

---

## What We Added This Sprint

Sprint 3 made the debugger actually useful. Four files changed significantly:

**`internal/dbgclient/dbgclient.go`** — the `Session` struct grew from a passive
connection holder into an active command sender. We added `SendCommand`, step methods
(`StepInto`, `StepOver`, `StepOut`, `Run`), and breakpoint methods
(`SetBreakpoint`, `RemoveBreakpoint`).

**`internal/source/source.go`** — went from an empty stub to a self-contained file
loader: `MapPath` translates Xdebug's container paths to host paths, `Format` reads
the file and produces tview-formatted output with line numbers and a highlight.

**`internal/breakpoints/breakpoints.go`** — a simple in-memory store for the active
breakpoint list with `Add`, `Remove`, and `Format`.

**`internal/tui/tui.go`** — the App struct gained a `sync.Mutex`, a stored session,
and a `handleCommand` method that dispatches all user input to the right DBGp call.

---

## Sending Commands: Writing to `net.Conn`

In Sprint 2 we only *read* from the connection. In Sprint 3 we write back.

**In PHP** you'd write to a socket with `fwrite($socket, $data)`.

**In Go:**
```go
func (s *Session) SendCommand(cmd string) error {
    s.txID++
    msg := fmt.Sprintf("%s -i %d\000", cmd, s.txID)
    _, err := s.conn.Write([]byte(msg))
    return err
}
```

`\000` is a null byte — the DBGp protocol uses it to terminate commands (just as
it uses it to frame messages). `fmt.Sprintf` builds the string, `[]byte(...)` converts
it, and `conn.Write` sends it. No flushing needed — `net.Conn` writes go straight to
the TCP stack.

The `s.txID++` is Go's equivalent of `$this->txId++` — post-increment on a struct field.

---

## Protecting Shared State: `sync.Mutex`

The `Session` is created inside the TCP goroutine but read from the UI goroutine
(when the user types a command). Two goroutines accessing the same pointer at the
same time is a data race — Go's race detector will flag it and it can cause crashes.

**The fix:** a `sync.Mutex` on the `App` struct.

```go
type App struct {
    mu      sync.Mutex
    session *dbgclient.Session
    // ...
}

func (a *App) SetSession(session *dbgclient.Session) {
    a.mu.Lock()
    a.session = session
    a.mu.Unlock()
}

func (a *App) getSession() *dbgclient.Session {
    a.mu.Lock()
    defer a.mu.Unlock()
    return a.session
}
```

**PHP equivalent:** In PHP you don't usually need this because PHP is single-threaded.
In Go, any time two goroutines share a variable, you need either a mutex or a channel.

`defer a.mu.Unlock()` is idiomatic Go — it guarantees the unlock happens when the
function returns, even if it returns early. PHP has no equivalent (though `finally`
is similar).

---

## Switch with Boolean Cases

PHP's `switch` matches a value. Go's `switch` can also work with boolean conditions —
like a chain of `if/elseif` but more readable:

**PHP:**
```php
if ($cmd === 's') {
    // step into
} elseif (str_starts_with($cmd, 'b ')) {
    // set breakpoint
} else {
    // unknown
}
```

**Go:**
```go
switch {
case cmd == "s":
    status, err = session.StepInto()
case strings.HasPrefix(cmd, "b "):
    // set breakpoint
default:
    a.SetStatus("unknown command: " + cmd)
}
```

`switch` with no expression is idiomatic Go for "match the first true case". It's
cleaner than a long `if/else if` chain. Note that Go's `switch` does **not**
fall through by default (unlike C/PHP) — each case is independent.

---

## Immediate Keypress Handling in tview

The `InputField` widget submits on Enter by default. For single-character debugger
commands (`s`, `n`, `o`, `r`) we want instant response — no Enter required.

tview's `SetInputCapture` on the `Application` intercepts every keypress globally:

```go
app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
    switch event.Rune() {
    case 'q':
        app.Stop()
        return nil  // consumed — don't pass to widgets
    case 's', 'n', 'o', 'r':
        if commandInput.GetText() == "" {
            go tuiApp.handleCommand(string(event.Rune()))
            return nil
        }
    }
    return event  // not consumed — pass through to focused widget
})
```

Returning `nil` consumes the event. Returning `event` passes it through.

The `commandInput.GetText() == ""` guard is important: if the user is typing
`b index.php:10`, pressing `n` while typing should add the letter `n` to the
input, not trigger a step-over.

---

## In-Memory Store: Slices as Collections

PHP developers reach for arrays naturally. In Go, a slice (`[]T`) is the equivalent.
A simple store pattern — keeping a slice on a struct and exposing Add/Remove methods —
is idiomatic for small collections:

```go
type Store struct {
    items []Breakpoint
}

func (s *Store) Add(file string, line int, id string) {
    s.items = append(s.items, Breakpoint{File: file, Line: line, ID: id})
}

func (s *Store) Remove(file string, line int) (id string, err error) {
    for i, bp := range s.items {
        if bp.File == file && bp.Line == line {
            s.items = append(s.items[:i], s.items[i+1:]...)
            return bp.ID, nil
        }
    }
    return "", fmt.Errorf("no breakpoint at %s:%d", file, line)
}
```

The removal trick `append(s[:i], s[i+1:]...)` splices out element `i` by
concatenating everything before it with everything after it. The `...` unpacks
the second slice as variadic arguments — like PHP's `...$array` spread operator.

---

## Input Validation with `strconv.Atoi`

A common pattern: check whether a string is a pure integer without caring about
the value. In PHP: `is_numeric($s)`. In Go:

```go
if _, numErr := strconv.Atoi(arg); numErr == nil {
    // arg is a pure integer
}
```

The `_` discards the parsed value — we only care whether parsing succeeded.
This is how `b 17` detects it's a bare line number and infers the filename
from the current session state.

---

## Shell Script Cleanup: `exec` vs Subprocess

The DDEV host command script originally used `exec ddev-xdebug-tui`, which
*replaces* the shell process with the binary. That's efficient but means no
cleanup code can run after the binary exits.

To auto-disable Xdebug when the debugger quits, we removed `exec`:

```bash
# Ensure Xdebug is enabled
ddev xdebug on

# Run as a subprocess (not exec) so cleanup below runs on exit
ddev-xdebug-tui

# Turn Xdebug back off when the debugger exits
ddev xdebug off
```

Without `exec`, the shell stays alive, waits for `ddev-xdebug-tui` to finish,
then runs `ddev xdebug off`. Simple and reliable.

---

## What We Built

At the end of Sprint 3, `ddev xdebug-tui` is a working interactive debugger:

- Xdebug auto-enables on launch and disables on exit
- PHP pauses at the first executable line immediately on browser visit
- Source panel shows the file with the current line highlighted in black-on-yellow
- `s` / `n` / `o` step through code; source panel updates after each step
- `r` runs to the next breakpoint
- `b index.php:16` or shorthand `b 16` sets a breakpoint
- `rb index.php:16` removes it
- Breakpoints panel lists all active breakpoints

Sprint 4 adds variable inspection and stack frame display.

![Sprint 3 final state](../WIREFRAMES/ddev-xdebug-tui-wireframe-s3-4.svg)
