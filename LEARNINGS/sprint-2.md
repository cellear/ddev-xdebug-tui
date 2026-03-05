# Sprint 2 Learnings: Go for PHP Developers

This document explains the Go concepts introduced in Sprint 2. If you haven't
read `LEARNINGS/sprint-1.md` yet, start there.

---

## What We Added This Sprint

Sprint 2 wired ddev-xdebug-tui to the real world. Three files changed:

**`internal/dbgclient/dbgclient.go`** — went from an empty stub to the heart
of the network layer. We added:
- `Listen()` — opens a TCP socket on port 9003 and waits for Xdebug to connect
- `ReadMessage()` — reads one complete DBGp message off the wire using the
  protocol's length-prefix framing
- `ParseInit()` — parses Xdebug's opening XML handshake to extract the language
  and the file being debugged

**`internal/tui/tui.go`** — gained two new methods for safely updating the UI
from background threads, and two panels (`statusBar`, `sourcePanel`) were
promoted from local variables to struct fields so they could be reached after
`NewApp()` returned.

**`cmd/ddev-xdebug-tui/main.go`** — grew from 12 lines to 54 lines. The main
function now spins up the TCP listener in a background goroutine, and the
callback it passes chains together: connect → read init packet → parse XML →
update status bar and source panel.

The result: visit the PHP test site in a browser with Xdebug on, and the TUI
reacts in real time.

---

## Goroutines — Go's Lightweight Threads

In PHP, everything runs sequentially — one request, one thread, top to bottom.
Go programs can do multiple things concurrently using **goroutines** — very
lightweight threads managed by the Go runtime (not the OS).

Starting a goroutine is just the keyword `go` before a function call. Here's
the actual code from `main.go`:

```go
// Start TCP listener in background goroutine.
// When Xdebug connects, read the init packet and display it.
go func() {
    err := dbgclient.Listen(func(conn net.Conn) {
        app.SetStatus("ddev-xdebug-tui | Xdebug connected")
        // ... read and parse init packet
    })
    if err != nil {
        app.SetStatus("ddev-xdebug-tui | listener error: " + err.Error())
    }
}()
```

The `func() { ... }()` part is an **immediately invoked anonymous function** —
a closure called right away. You see this pattern constantly with goroutines.
The `go` keyword is the only thing that makes it concurrent; without it, this
would just be a regular blocking function call.

While the goroutine sits blocked waiting for Xdebug to connect, the main
goroutine keeps the TUI responsive.

**The important rule:** goroutines run concurrently, which means two goroutines
can try to modify the same data at the same time. This causes **race conditions**
— bugs that are very hard to reproduce. Which brings us to...

---

## QueueUpdateDraw — Thread-Safe UI Updates

When our goroutine receives an Xdebug connection, it wants to update the TUI.
But tview is not thread-safe — you can't update UI elements from a goroutine
directly. Doing so causes a race condition and may crash.

The solution is `app.QueueUpdateDraw()`. Here's the actual `SetStatus` method
we added to `tui.go`:

```go
// SetStatus updates the status bar text. Safe to call from any goroutine.
// Uses QueueUpdateDraw to avoid race conditions when called from background threads.
func (a *App) SetStatus(text string) {
    a.app.QueueUpdateDraw(func() {
        a.statusBar.SetText(text)
    })
}
```

`QueueUpdateDraw` schedules the update on the main UI goroutine rather than
running it immediately. Think of it like JavaScript's `setTimeout(fn, 0)` —
you're not doing the work now, you're handing it off to the right thread to do
safely on the next draw cycle.

Notice that `statusBar` is now a field on the `App` struct (not a local
variable inside `NewApp`). That change was necessary so `SetStatus` could
reach it after construction. This is a common Go pattern: promote a value
from local scope to a struct field when you need to access it later.

---

## The `net` Package — TCP in Go

Go's standard library handles TCP networking directly with no extension needed.
In `dbgclient.go`:

```go
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

    firstConn, err := listener.Accept()
    if err != nil {
        return err
    }
    onConnect(firstConn)
    // ...
}
```

`net.Listen` opens the socket (PHP: `stream_socket_server`).
`listener.Accept()` blocks until a client connects (PHP: `stream_socket_accept`).
`net.Conn` is the connection object — you read from it and write to it like a file.

`defer listener.Close()` is worth noting: `defer` schedules a call to run when
the surrounding function returns, no matter how it returns (normal exit, error,
panic). It's the Go equivalent of PHP's `finally` block. You'll see `defer`
used constantly for cleanup.

---

## Function Values — Passing Functions as Arguments

Go functions are values and can be passed as arguments. The `Listen` signature:

```go
func Listen(onConnect func(net.Conn)) error
```

`func(net.Conn)` is a type — "a function that accepts a `net.Conn`". In `main.go`
we pass a closure that chains the whole init sequence together:

```go
dbgclient.Listen(func(conn net.Conn) {
    app.SetStatus("ddev-xdebug-tui | Xdebug connected")

    data, err := dbgclient.ReadMessage(conn)
    if err != nil {
        app.SetStatus("ddev-xdebug-tui | read error: " + err.Error())
        conn.Close()
        return
    }

    language, fileURI, err := dbgclient.ParseInit(data)
    if err != nil {
        app.SetStatus("ddev-xdebug-tui | parse error: " + err.Error())
        conn.Close()
        return
    }

    app.SetInitInfo(language, fileURI)
    // ...
})
```

This is how Go achieves callbacks without needing interfaces or abstract classes.
Each `if err != nil` block is Go's equivalent of a `catch` — more on that in
the error handling section below.

---

## `bufio.Reader` and `io.ReadFull` — Reading Off the Wire

The DBGp protocol frames messages as `<decimal-length>\0<xml-payload>\0`. We
need to read the length prefix one byte at a time (we don't know where it ends),
then read the payload in one efficient bulk read. Here's the actual
`ReadMessage` implementation:

```go
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
        return nil, fmt.Errorf("error reading final null terminator: %w", err)
    }
    if b != 0 {
        return nil, fmt.Errorf("expected null terminator, got: %c", b)
    }

    return payload, nil
}
```

A few Go-specific things happening here:

**`bufio.NewReader`** wraps the connection with buffering. Without it, each
`ReadByte()` would make a separate system call (slow). With buffering, Go reads
a chunk into memory and serves bytes from that buffer.

**`make([]byte, length)`** allocates a byte slice of exactly `length` bytes,
all initialised to zero. The PHP equivalent is roughly `str_repeat("\0", $length)`.
`[]byte` is Go's type for a mutable sequence of bytes — distinct from `string`,
which is immutable. You'll see `make` everywhere in Go for allocating slices and maps.

**`io.ReadFull`** reads until the buffer is completely filled. A plain
`reader.Read(payload)` might return fewer bytes than requested — TCP delivers
data in chunks and Read returns whatever arrived. `ReadFull` keeps looping
until it has everything or hits an error.

---

## XML Parsing with `encoding/xml`

Go parses XML by mapping it onto structs using **struct tags** — metadata in
backtick strings. From `dbgclient.go`:

```go
type initPacket struct {
    XMLName  xml.Name `xml:"init"`
    Language string   `xml:"language,attr"`
    FileURI  string   `xml:"fileuri,attr"`
}

func ParseInit(data []byte) (language string, fileURI string, err error) {
    // Go's xml package only supports UTF-8. Xdebug declares iso-8859-1 but
    // the content is ASCII-compatible, so we rewrite the declaration before parsing.
    data = bytes.ReplaceAll(data, []byte(`encoding="iso-8859-1"`), []byte(`encoding="UTF-8"`))

    var packet initPacket
    err = xml.Unmarshal(data, &packet)
    if err != nil {
        return "", "", fmt.Errorf("failed to parse init packet: %w", err)
    }
    // ...
    return packet.Language, packet.FileURI, nil
}
```

The struct tags:
- `` `xml:"init"` `` — this struct maps to an `<init>` element
- `` `xml:"language,attr"` `` — maps to the `language` *attribute*
- `` `xml:"fileuri,attr"` `` — maps to the `fileuri` *attribute*

In PHP: `simplexml_load_string($xml)->attributes()['language']`. Go's approach
is more verbose but the mapping is explicit and checked at compile time.

### The iso-8859-1 Gotcha

Xdebug sends `<?xml version="1.0" encoding="iso-8859-1"?>`. Go's `encoding/xml`
only accepts UTF-8. The content is ASCII-compatible, so the fix is a single
substitution before parsing — that's the `bytes.ReplaceAll` line above. This is
the kind of real-world protocol detail that no documentation warns you about.

---

## Error Handling — Multiple Return Values

Sprint 1 introduced `if err != nil`. Sprint 2 shows it at scale. Go functions
can return multiple values, and the convention is to return `(result, error)`:

```go
language, fileURI, err := dbgclient.ParseInit(data)
if err != nil {
    app.SetStatus("ddev-xdebug-tui | parse error: " + err.Error())
    conn.Close()
    return
}
```

Each call that can fail gets its own `if err != nil` check immediately after.
There's no try/catch block wrapping a whole section — errors are handled at
the point they occur, explicitly. It's more verbose than PHP exceptions but
makes the error paths visible in the code rather than hidden in a catch block
elsewhere.

The `%w` verb in `fmt.Errorf("failed to parse: %w", err)` wraps the original
error inside the new one, like PHP's exception chaining
(`new RuntimeException("context", 0, $previous)`).

---

## `strings.LastIndex` — Slicing Strings

To get just `index.php` from `file:///var/www/html/index.php`, from `main.go`:

```go
filename := fileURI
if idx := strings.LastIndex(fileURI, "/"); idx >= 0 {
    filename = fileURI[idx+1:]
}
app.SetStatus(fmt.Sprintf("ddev-xdebug-tui | %s | %s", language, filename))
```

`strings.LastIndex` returns the position of the last `/`. Then `fileURI[idx+1:]`
takes everything after it — Go's **slice notation**. `s[start:]` means "from
`start` to the end of the string." PHP equivalent: `substr($s, $idx + 1)`.

`fmt.Sprintf` works exactly like PHP's `sprintf` — same `%s` placeholders,
same idea.

---

## What We Built

At the end of Sprint 2, visiting the PHP test site in a browser with Xdebug
enabled causes the TUI to:

1. Accept the incoming Xdebug TCP connection on port 9003
2. Update the status bar: `"ddev-xdebug-tui | Xdebug connected"`
3. Read the first DBGp message using length-prefix framing
4. Parse the `<init>` XML packet to extract language and file URI
5. Update the Source panel: `Language: PHP` / `File: file:///var/www/html/index.php`
6. Update the status bar: `"ddev-xdebug-tui | PHP | index.php"`

Sprint 3 adds the ability to send DBGp commands back to Xdebug — enabling
stepping, breakpoints, and variable inspection.

![Sprint 2 final state](../WIREFRAMES/ddev-xdebug-tui-wireframe-s2-4.svg)
