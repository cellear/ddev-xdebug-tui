# Sprint 2: TCP Connection + DBGp Init

**Sprint Goal:** When `ddev xdebug-tui` is running and the developer visits the
PHP test site in a browser with Xdebug enabled, the TUI reacts visibly — first
by showing "Xdebug connected", then by displaying the file URI and language from
Xdebug's init packet.

---

## Demo Checkpoints

This sprint has two demo points:

**Demo A — after S2-2:**
1. Run `ddev xdebug-tui`
2. Visit `https://php-test-project.ddev.site` in browser
3. Status bar updates: "Xdebug connected"

**Demo B — after S2-4:**
1. Same as above
2. Source panel (or status bar) shows the fileURI and language from Xdebug's init packet

---

## Stories

---

### S2-1: TCP Listener
**Status:** [done]
**Owner:** claude-haiku-4-5

**Description:**
Implement a TCP listener inside `internal/dbgclient/` that listens on port 9003
and accepts one connection. Per `AGENT.md`, single-session only — additional
incoming connections are ignored. The listener runs in a goroutine and notifies
the TUI via a callback when a client connects.

**Acceptance Criteria:**
- `internal/dbgclient/dbgclient.go` implements a `Listen(onConnect func(net.Conn))` function
- Listener binds to `:9003` on startup
- Only the first connection is accepted; subsequent connections are closed immediately
- The goroutine is minimal and clearly commented (per `AGENT.md` concurrency policy)
- `go build ./...` succeeds

**Notes:**
No UI changes in this story — the callback wires up in S2-2. This story is
purely the network layer.

---

### S2-2: Wire Connection Status to TUI
**Status:** [done]
**Owner:** claude-haiku-4-5

**Prerequisites:** S2-1 done and compiling.

**Description:**
Start the TCP listener from `main.go` (or `tui`) and wire the `onConnect`
callback so the status bar updates when Xdebug connects. This is the first
story with a visible demo.

**Acceptance Criteria:**
- `dbgclient.Listen(...)` is called at startup
- When Xdebug connects, the status bar text changes to:
  `"ddev-xdebug-tui | Xdebug connected"`
- The TUI remains responsive while waiting (no blocking the main thread)
- `go build ./...` succeeds

**Notes:**
tview UI updates from goroutines must use `app.QueueUpdateDraw(func() {...})` —
direct updates from a goroutine will cause a race condition. Make sure the Haiku
prompt explicitly mentions this.

**→ Demo A checkpoint after this story.**

---

### S2-3: DBGp Message Framing
**Status:** [backlog]
**Owner:** claude-sonnet-4-6

**Prerequisites:** S2-2 done and Demo A passed.

**Description:**
Implement correct DBGp message framing in `internal/dbgclient/`. DBGp messages
are length-prefixed: a decimal byte count, a null byte, the XML payload, another
null byte. The parser must handle partial reads correctly.

**DBGp framing format:**
```
<length>\0<xml-payload>\0
```
Example:
```
148\0<?xml version="1.0"...?><init .../>  \0
```

**Acceptance Criteria:**
- `ReadMessage(conn net.Conn) ([]byte, error)` correctly reads one full DBGp message
- Handles partial TCP reads (reads until null terminator, not fixed buffer)
- Returns the raw XML bytes for the caller to parse
- Handles EOF and connection errors gracefully (returns error, doesn't panic)
- Unit test or inline test demonstrating framing with a mock reader
- `go build ./...` succeeds

**Notes:**
This is the most protocol-sensitive code in Sprint 2. Take care with the null
byte terminators — a common mistake is treating the length prefix as fixed-width
rather than reading until `\0`.

---

### S2-4: Parse Init Packet + Display in TUI
**Status:** [backlog]
**Owner:** claude-sonnet-4-6

**Prerequisites:** S2-3 done and compiling.

**Description:**
Parse the Xdebug `<init>` packet (the first message Xdebug sends on connect)
and display the fileURI and language in the TUI. This confirms the full pipeline:
TCP connect → frame read → XML parse → UI update.

**The init packet looks like:**
```xml
<?xml version="1.0" encoding="iso-8859-1"?>
<init xmlns="urn:debugger_protocol_v1"
      language="PHP"
      protocol_version="1.0"
      fileuri="file:///var/www/html/index.php"
      appid="..."
      idekey="..."/>
```

**Acceptance Criteria:**
- `<init>` XML is parsed using Go's standard `encoding/xml` package (no extra deps)
- `language` and `fileuri` attributes are extracted
- Source panel updates to show:
  ```
  Language: PHP
  File:     file:///var/www/html/index.php
  ```
- Status bar updates to: `"ddev-xdebug-tui | PHP | index.php"`
  (just the filename, not the full URI, in the status bar)
- `go build ./...` succeeds

**→ Demo B checkpoint after this story.**

---

## Sprint Review Demo Checklist

**Demo A:**
1. `ddev xdebug-tui` launches (status: "waiting for Xdebug connection")
2. Visit `https://php-test-project.ddev.site` in browser
3. Status bar updates to "Xdebug connected"
4. Press `q` to exit

**Demo B:**
1. Same as Demo A steps 1–2
2. Status bar shows: `"ddev-xdebug-tui | PHP | index.php"`
3. Source panel shows language and fileURI
4. Press `q` to exit

---

## Decisions Made

- **Port:** 9003 (Xdebug default)
- **Single session only:** additional connections closed immediately
- **XML parsing:** Go standard library `encoding/xml` — no new dependencies
- **TUI update pattern:** `app.QueueUpdateDraw()` for all updates from goroutines

---

## Deferred to Later Sprints

- Sending DBGp commands back to Xdebug (Sprint 3)
- Step controls (Sprint 3)
- Breakpoints (Sprint 3/4)
- Path mapping: container paths (`/var/www/html/...`) vs host paths (Sprint 3 — flag when implementing source loading)

---

Last updated: 2026-03-05 by claude-haiku-4-5
