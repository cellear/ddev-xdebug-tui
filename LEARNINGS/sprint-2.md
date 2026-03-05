# Sprint 2 Learnings: Go for PHP Developers

This document explains the Go concepts introduced in Sprint 2, which added a
TCP listener, Xdebug connection handling, DBGp message framing, and init packet
parsing. If you haven't read `LEARNINGS/sprint-1.md` yet, start there.

---

## Goroutines — Go's Lightweight Threads

In PHP, everything runs sequentially. One request, one thread, top to bottom.
Go programs can run multiple things concurrently using **goroutines** — very
lightweight threads managed by the Go runtime (not the OS).

Starting a goroutine is just the keyword `go` before a function call:

```go
go func() {
    // this runs concurrently with everything else
    dbgclient.Listen(...)
}()
```

The `func() { ... }()` part is an **immediately invoked anonymous function** —
the same concept as PHP's `(function() { ... })()` in JavaScript, or a closure
called right away. In Go you see this pattern constantly with goroutines.

Our TCP listener runs in a goroutine so it can wait for Xdebug to connect
without freezing the TUI. While the goroutine sits blocked on `Accept()`,
the main goroutine keeps the UI responsive.

**The important rule:** goroutines run concurrently, which means two goroutines
can try to modify the same data at the same time. This causes **race conditions**
— bugs that are very hard to reproduce. Go has a built-in race detector
(`go run -race ./...`) that can catch these.

---

## The `net` Package — TCP in Go

Go's standard library handles TCP networking directly, no extension needed.
PHP developers are used to sockets being somewhat painful. In Go they're
a first-class citizen:

```go
// PHP equivalent: stream_socket_server('tcp://0.0.0.0:9003', ...)
listener, err := net.Listen("tcp", ":9003")

// PHP equivalent: stream_socket_accept($server)
conn, err := listener.Accept()
```

`net.Conn` is an interface (more on interfaces in a future sprint) that
represents any network connection. You can read from it and write to it like
a file. `Accept()` blocks until a client connects — which is exactly why
we run it in a goroutine.

---

## Function Values — Passing Functions as Arguments

In PHP you pass callbacks using `callable` or closures:

```php
array_map(function($x) { return $x * 2; }, $items);
```

Go has the same concept — functions are values and can be passed as arguments.
The type of a function is written as its signature:

```go
// A function that takes a net.Conn and returns nothing
func Listen(onConnect func(net.Conn)) error
```

`func(net.Conn)` is the type — "a function that accepts a net.Conn". When we
call `Listen`, we pass in a closure:

```go
dbgclient.Listen(func(conn net.Conn) {
    app.SetStatus("Xdebug connected")
})
```

This is how Go achieves callbacks without needing interfaces or abstract classes.

---

## QueueUpdateDraw — Thread-Safe UI Updates

Here's one of the trickiest things in Sprint 2. When our goroutine receives
an Xdebug connection, it wants to update the TUI (change the status bar text).
But tview is not thread-safe — you can't update UI elements from a goroutine
directly. Doing so causes a race condition.

The solution is `app.QueueUpdateDraw()`:

```go
// WRONG — race condition, may crash:
a.statusBar.SetText("Xdebug connected")

// CORRECT — schedules the update on the main UI thread:
a.app.QueueUpdateDraw(func() {
    a.statusBar.SetText("Xdebug connected")
})
```

`QueueUpdateDraw` puts the function into a queue that the main goroutine
processes on the next draw cycle. Think of it like JavaScript's
`setTimeout(fn, 0)` — you're not doing the work now, you're scheduling it
to happen safely on the right thread.

The PHP world doesn't usually deal with this because PHP-FPM handles one
request per process. This is one of the genuine mental shifts when coming
to Go.

---

## `bufio.Reader` — Buffered Reading

Go's `net.Conn` lets you read bytes from a network connection, but raw reads
are painful for protocols like DBGp where you need to read one byte at a time
to find a delimiter.

`bufio.Reader` wraps any reader and adds buffering and convenience methods:

```go
reader := bufio.NewReader(conn)

// Read one byte at a time efficiently
b, err := reader.ReadByte()
```

Without buffering, each `ReadByte()` would make a separate system call (slow).
With `bufio.Reader`, Go reads a chunk into memory and serves bytes from that
buffer — much faster.

PHP's equivalent is `fread()` vs `stream_get_contents()` — you can read byte
by byte, but the stream already buffers under the hood.

---

## `io.ReadFull` — Reading Exactly N Bytes

Once we know the DBGp message length, we need to read exactly that many bytes.
`io.ReadFull` does this reliably:

```go
payload := make([]byte, length)  // allocate a slice of exactly `length` bytes
n, err := io.ReadFull(reader, payload)
```

Without `ReadFull`, a plain `reader.Read(payload)` might return fewer bytes
than requested (TCP can deliver data in chunks). `ReadFull` keeps reading until
the buffer is completely filled or an error occurs.

This is the kind of thing PHP abstracts away. In Go you're closer to the
metal — and `io.ReadFull` is the standard solution.

---

## `make` — Allocating Slices

You'll see `make` all over Go code. For slices (Go's arrays):

```go
payload := make([]byte, length)
```

This creates a slice of `length` bytes, all initialised to zero. The PHP
equivalent is roughly `str_repeat("\0", $length)` — a string of null bytes.

In Go, `[]byte` is a slice of bytes. Strings and byte slices are related but
distinct; you can convert between them with `string(bytes)` and `[]byte(str)`.

---

## XML Parsing with `encoding/xml`

Go parses XML by mapping it onto structs using **struct tags** — metadata
annotations in backtick strings:

```go
type initPacket struct {
    XMLName  xml.Name `xml:"init"`
    Language string   `xml:"language,attr"`
    FileURI  string   `xml:"fileuri,attr"`
}
```

The backtick strings (`` `xml:"..."` ``) tell the XML parser:
- `xml:"init"` — this struct maps to an element named `<init>`
- `xml:"language,attr"` — this field maps to the `language` *attribute*
- `xml:"fileuri,attr"` — this field maps to the `fileuri` *attribute*

Then you unmarshal:

```go
var packet initPacket
err = xml.Unmarshal(data, &packet)
// packet.Language is now "PHP"
// packet.FileURI is now "file:///var/www/html/index.php"
```

In PHP this would be `simplexml_load_string($xml)->attributes()['language']`.
Go's approach is more verbose but catches type mismatches at compile time.

### The iso-8859-1 Gotcha

Xdebug sends this XML declaration:

```xml
<?xml version="1.0" encoding="iso-8859-1"?>
```

Go's `encoding/xml` only supports UTF-8. Since the actual content is
ASCII-compatible, the fix is a simple substitution before parsing:

```go
data = bytes.ReplaceAll(
    data,
    []byte(`encoding="iso-8859-1"`),
    []byte(`encoding="UTF-8"`),
)
```

This is a real-world example of a protocol implementation detail that no
documentation will warn you about — you find it by hitting the error
(`xml: encoding "iso-8859-1"`) and reasoning about the fix.

---

## Error Wrapping with `%w`

In Sprint 1 we saw basic error handling. Sprint 2 introduced **error wrapping**:

```go
return "", "", fmt.Errorf("failed to parse init packet: %w", err)
```

The `%w` verb (not `%v` or `%s`) wraps the original error inside a new one.
This preserves the original error so callers can inspect it with
`errors.Is()` or `errors.As()` — like PHP's exception chaining
(`new RuntimeException("context", 0, $previous)`).

For now just know that `%w` is the right choice when wrapping errors, and
`%v` is for when you just want to include the error message as a string.

---

## `strings.LastIndex` — Finding the Last Occurrence

To extract just the filename from `file:///var/www/html/index.php`:

```go
if idx := strings.LastIndex(fileURI, "/"); idx >= 0 {
    filename = fileURI[idx+1:]
}
```

`strings.LastIndex` finds the position of the last `/`. Then `fileURI[idx+1:]`
takes everything after it — Go's **slice notation** for strings.

`s[start:end]` gives a substring. Omitting `end` means "to the end of the
string". The PHP equivalent is `substr($s, $idx + 1)`.

---

## What We Built

At the end of Sprint 2, visiting the PHP test site in a browser with Xdebug
enabled causes the TUI to:

1. Detect the incoming Xdebug connection on port 9003
2. Update the status bar: `"ddev-xdebug-tui | Xdebug connected"`
3. Read and frame the first DBGp message (the init packet)
4. Parse the XML to extract language and file URI
5. Display in the Source panel: `Language: PHP` / `File: file:///...`
6. Update the status bar: `"ddev-xdebug-tui | PHP | index.php"`

Sprint 3 adds the ability to send DBGp commands back to Xdebug — enabling
stepping, breakpoints, and variable inspection.
