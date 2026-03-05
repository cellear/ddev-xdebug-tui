# Sprint 4 — Variables, Stack, Restart, Polish

**Goal:** Complete the PoC. Populate the two empty panels (variables, stack), add
session restart, and run the full acceptance flow.

**Demo:** Full manual acceptance run — start → breakpoint → browser request → step
→ inspect variables & stack → script finishes → auto-restart → second request →
quit.

---

## Stories

### S4-1 · Variables panel · [done]

**Owner:** Haiku
**Scope:** Call `context_get` after every step/break, parse the DBGp XML response,
display variable name/value pairs in the Variables panel.

**Acceptance criteria:**
- After each step, Variables panel shows local variables for the current frame
- Format: `$name = value` one per line
- Objects/arrays collapsed to `$name = {object}` / `$name = [array]` (no deep expand yet)
- Panel cleared and repopulated on each step

**DBGp command:**
```
context_get -i N -d 0
```
(`-d 0` = depth 0 = current frame locals)

**XML response shape:**
```xml
<response command="context_get" context="0" transaction_id="N">
  <property name="$foo" fullname="$foo" type="string" size="3">
    <![CDATA[bar]]>
  </property>
  <property name="$n" fullname="$n" type="int">42</property>
  <property name="$obj" fullname="$obj" type="object" classname="Foo" children="1" numchildren="2"/>
</response>
```

---

### S4-2 · Stack panel · [done]

**Owner:** Haiku
**Scope:** Call `stack_get` after every step/break, parse the DBGp XML response,
display stack frames in the Stack panel.

**Acceptance criteria:**
- After each step, Stack panel shows one line per frame
- Format: `► file.php:line` for current frame (depth 0), `  file.php:line` for others
- Filename only (not full path)
- Panel cleared and repopulated on each step

**DBGp command:**
```
stack_get -i N
```

**XML response shape:**
```xml
<response command="stack_get" transaction_id="N">
  <stack level="0" type="file" filename="file:///var/www/html/index.php" lineno="17" where="{main}"/>
  <stack level="1" type="file" filename="file:///var/www/html/lib/greeter.php" lineno="5" where="greet"/>
</response>
```

---

### S4-3 · Session restart · [done]

**Owner:** Haiku
**Scope:** After receiving `status=stopping` (or `status=stopped`) from any step
response, update status bar to "Script finished — waiting for next connection…",
leave screen content as-is, and loop the TCP listener to accept the next connection.

**Acceptance criteria:**
- Status bar updates immediately on stopping/stopped
- Source, variables, stack panels stay visible (not cleared)
- Listener accepts next connection without user restarting the binary
- On new connection: session replaced, source/variables/stack refresh from new session

**Implementation note:** The TCP accept loop is in `main.go`. Currently it accepts
once and blocks. Change to a `for` loop: after session ends, call `app.ClearSession()`
(sets status, nils session) then loop back to `ln.Accept()`.

---

### S4-4 · `rb N` shorthand · [done]

**Owner:** Haiku
**Scope:** Mirror the `b N` shorthand for remove-breakpoint. `rb 17` removes the
breakpoint at line 17 of the current file, inferring filename from `session.CurrentFile`.

**Acceptance criteria:**
- `rb 17` works identically to `rb index.php:17` when current file is index.php
- Error message if no breakpoint at that line

---

## Demo · Full Acceptance Run

```
ddev xdebug-tui
> b 10          (set breakpoint at current file line 10)
[browser] curl http://my-project.ddev.site/
[TUI pauses at line 10]
[Variables panel shows locals]
[Stack panel shows call stack]
> n             (step over)
> n             (step over)
> r             (run to next breakpoint / end)
[status: Script finished — waiting for next connection…]
[browser] curl http://my-project.ddev.site/
[TUI pauses again at line 10]
> q             (quit)
```

**Demo status: [ ] PENDING**

---

## Wireframes

_To be added after implementation._

---

## Notes

- `context_get` and `stack_get` should be called together after every step response,
  before refreshing the UI — one `QueueUpdateDraw` at the end covers all panels.
- Path mapping for stack frames: strip `file://` + apply same `MapPath` used for source.
- Filename-only display: `filepath.Base(hostPath)`.
- Session restart: `app.ClearSession()` should set status bar text and nil the session
  pointer under mutex. Variables/stack/source panels intentionally NOT cleared.
