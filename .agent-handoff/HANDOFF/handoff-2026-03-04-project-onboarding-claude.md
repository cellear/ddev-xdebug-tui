# Handoff — 2026-03-04 — Project Onboarding — Claude

## What Happened This Session

- Read all project docs: `AGENT.md`, `README.md`, `REPO_BOOTSTRAP.md`, `DOC/implementation-plan.md`
- Confirmed: zero Go code exists yet — project is pre-implementation
- Discussed session continuity: handoff protocol is the right tool across interfaces

## Current State

**No code has been written.** Repository contains only docs and planning files.

## What To Do Next

Start Step 1 from `REPO_BOOTSTRAP.md`:

1. Create directory structure:
   - `cmd/ddev-xdebug-tui/main.go`
   - `internal/dbgclient/`
   - `internal/session/`
   - `internal/breakpoints/`
   - `internal/source/`
   - `internal/tui/`

2. Run: `go mod init github.com/<username>/ddev-xdebug-tui`

3. Add dependencies:
   - `github.com/rivo/tview`
   - `github.com/gdamore/tcell/v2`

4. Write a minimal `main.go` that starts a tview app showing:
   ```
   ddev-xdebug-tui
   waiting for Xdebug connection
   ```

Each step must compile and run before moving to the next.

## Key Constraints (from AGENT.md)

- Language: Go
- UI: tview only
- One connection at a time
- No conditional breakpoints, no multi-session, no heavy frameworks
- Keep code small, readable, explicit
- Maintainer reads every line before release

## Files Created or Modified

None — read-only session.

## References

- `AGENT.md` — core constraints and philosophy
- `REPO_BOOTSTRAP.md` — 13-step incremental build plan
- `DOC/implementation-plan.md` — token-efficient plan (last updated 2026-03-04 by Codex)
