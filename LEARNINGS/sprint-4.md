# LEARNINGS — Sprint 4

> Concepts encountered during Sprint 4. Written for PHP developers learning Go.

---

## 1. Channel-based lifecycle signalling (`chan struct{}`)

We needed `main.go` to know when a debug session ended (script finished) so it
could loop back and accept the next connection. Rather than polling or using a
boolean flag with a mutex, we used a **done channel**:

```go
type Session struct {
    Done      chan struct{}
    closeOnce sync.Once
}

func (s *Session) Close() {
    s.closeOnce.Do(func() {
        close(s.Done)
    })
}
```

In `main.go`:

```go
<-session.Done  // blocks until Close() is called
app.ClearSession()
// loop back to Accept()
```

`close(ch)` on a `chan struct{}` unblocks all goroutines waiting on `<-ch`. It
costs no memory (zero-width type) and is the idiomatic Go signal for "this thing
is done." Think of it like a PHP event that multiple listeners can subscribe to,
except in Go it's built into the language.

---

## 2. `sync.Once` for idempotent operations

`close()` on an already-closed channel panics. If `session.Close()` were called
twice (e.g. by `handleCommand` and some error path), the program would crash.
`sync.Once` guarantees the closure runs exactly once no matter how many goroutines
call `Close()`:

```go
s.closeOnce.Do(func() {
    close(s.Done)
})
```

In PHP terms: `sync.Once` is like a flag variable guarding a block — `if (!$done) { $done = true; ... }` — except it's thread-safe by construction.

---

## 3. Sequential TCP accept loop

In Sprint 3, `Listen` accepted one connection and closed all others. For
auto-restart we changed it to loop sequentially: accept → handle → accept next.

```go
func Listen(onConnect func(net.Conn)) error {
    ln, _ := net.Listen("tcp", ":9003")
    defer ln.Close()
    for {
        conn, err := ln.Accept()
        if err != nil { return err }
        onConnect(conn) // blocks until session ends
    }
}
```

The key insight: `onConnect` must **block** for the loop to be sequential. If it
returned immediately (like a fire-and-forget callback), the loop would race to
accept the next connection before the first session was ready.

---

## 4. `encoding/base64` — Xdebug encodes string values

Xdebug base64-encodes string property values in `context_get` XML responses:

```xml
<property name="$name" type="string" encoding="base64">
    <![CDATA[d29ybGQ=]]>
</property>
```

`d29ybGQ=` decodes to `world`. Integers, booleans, objects, and arrays are NOT
encoded — only strings. The `encoding="base64"` attribute tells you when to decode.

```go
import "encoding/base64"

if p.Encoding == "base64" {
    if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil {
        raw = string(decoded)
    }
}
```

`base64.StdEncoding.DecodeString` returns `([]byte, error)`. Convert to string
with `string(decoded)`. In PHP: `base64_decode($value)`.

---

## 5. XML `chardata` vs attributes

In Go's `encoding/xml`, `xml:",chardata"` reads the text content of an element
(including CDATA sections):

```go
type contextProperty struct {
    Name     string `xml:"name,attr"`    // reads the "name" attribute
    Value    string `xml:",chardata"`    // reads the text/CDATA content
}
```

For `<property name="$x"><![CDATA[hello]]></property>`:
- `Name` = `"$x"` (attribute)
- `Value` = `"hello"` (chardata)

The `,chardata` tag works on CDATA too — Go's XML parser treats them the same.

---

## 6. Preserving UI state on session end

When the PHP script finishes, the user should still see the last source/variables/
stack state (useful for post-mortem inspection). We achieved this by:

- `ClearSession()` nils the session pointer but does **not** clear any panels
- The listener loops to `Accept()` — a new connection will overwrite panels
- The status bar updates to "Script finished — waiting for next connection…"

The panels are `*tview.TextView` — they hold their own state. As long as we
don't call `SetText("")` on them, their content persists indefinitely.

---

## 7. Debugging "wrong binary" issues

When `go build -o bin/foo ./cmd/foo/` succeeds silently, the binary is in
`bin/foo` relative to the repo. But if `which foo` points somewhere else
(e.g. `~/go/bin/foo` from a previous `go install`), the new build is never run.

Solutions:
- `go install ./cmd/foo/` — builds AND installs to `$GOPATH/bin` in one step
- `cp bin/foo $(which foo)` — manual copy to the installed location

For visibility: put the version in a UI element that is **always visible during
an active session** (like a panel title), not in text that gets overwritten
(like a status bar that changes on connect).

---

## 8. `currentFileBase()` — DRY helper for filename extraction

We had the same "strip directory prefix from container URI" pattern in three
places. Extract it:

```go
func currentFileBase(session *dbgclient.Session) string {
    f := session.CurrentFile
    if idx := strings.LastIndex(f, "/"); idx >= 0 {
        f = f[idx+1:]
    }
    return f
}
```

`strings.LastIndex` finds the last `/` — correct for both Unix paths and
`file:///var/www/html/index.php` URIs. In PHP: `basename($path)`.

Last updated: 2026-03-05 by claude-sonnet-4-6
