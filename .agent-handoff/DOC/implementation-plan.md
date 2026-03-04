# Implementation Plan (Token-Efficient)

Goal: Deliver a minimal PoC for `ddev-xdebug-tui` that matches `AGENT.md` scope while minimizing high-cost model usage.

## Step 1: Repository Bootstrap and Guardrails
- Outcome: Create `go.mod`, `cmd/ddev-xdebug-tui/main.go`, and `internal/{dbgclient,session,breakpoints,source,tui}` skeleton.
- Primary owner: Human or low-reasoning model.
- High-reasoning needed: No.
- Token strategy: Use a short checklist and simple file scaffolding; avoid long design discussion.
- Human participation option: Human can create folders/files and run `go mod init` + dependency adds directly.

## Step 2: Minimal Runnable TUI Shell
- Outcome: App starts and shows “waiting for Xdebug connection”.
- Primary owner: Low-reasoning model.
- High-reasoning needed: No.
- Token strategy: Keep to one `main.go` pass; defer architecture debates.
- Human participation option: Human can verify terminal rendering and basic startup manually.

## Step 3: TCP Listener + First Connection Attachment
- Outcome: Listen on `:9003`, accept first connection only, show connection status in UI.
- Primary owner: Low-to-mid model.
- High-reasoning needed: Usually no; only if event flow gets tricky.
- Token strategy: Keep single-session policy explicit; reject/ignore extra sessions with a simple message.
- Human participation option: Human can trigger a browser request with Xdebug enabled to confirm connect behavior.

## Step 4: DBGp Framing + Minimal Command Transport
- Outcome: Read DBGp message framing reliably and send basic commands with transaction IDs.
- Primary owner: High-reasoning model recommended.
- High-reasoning needed: Yes (protocol edge cases, framing correctness).
- Token strategy: Narrow scope to supported command subset only; no broad protocol implementation.
- Human participation option: Human can run capture-style manual tests and report malformed-message cases.

## Step 5: Session State + Step Commands
- Outcome: Track current file/line/stack summary and wire `step_into`, `step_over`, `step_out`, `run`.
- Primary owner: Mid model.
- High-reasoning needed: No, unless state transitions become inconsistent.
- Token strategy: Keep state struct small and explicit; avoid abstractions.
- Human participation option: Human can execute stepping loop and note any incorrect pause/continue behavior.

## Step 6: Breakpoints (Ephemeral, Line-Only)
- Outcome: In-memory breakpoint store + `b file.php:line` / `rb file.php:line` mapped to `breakpoint_set` and `breakpoint_remove`.
- Primary owner: Mid model.
- High-reasoning needed: No.
- Token strategy: Parse one command format; skip conditionals/persistence.
- Human participation option: Human can enter breakpoint commands manually and validate behavior quickly.

## Step 7: Stack/Variables/Source Panels
- Outcome: Populate split-pane UI with stack, source context, variables, and breakpoint list.
- Primary owner: Mid-to-high model (UI integration + data mapping).
- High-reasoning needed: Sometimes yes (if DBGp payload mapping is ambiguous).
- Token strategy: Keep visual structure fixed; no UI feature expansion.
- Human participation option: Human can evaluate readability/usability and request small layout tweaks.

## Step 8: DDEV Command Wiring + Manual Acceptance Loop
- Outcome: `ddev debug` launches tool; full manual flow passes:
  1. start debugger
  2. set breakpoint
  3. trigger request
  4. step
  5. inspect variables
  6. quit
- Primary owner: Human + low/mid model.
- High-reasoning needed: No, unless environment-specific integration fails.
- Token strategy: Prefer concise runbook updates over long troubleshooting narratives.
- Human participation option: Human should drive real environment verification and provide precise failure notes between model runs.

## Model Usage Policy by Phase
- Prefer low/mid model for scaffolding, config, command wiring, and straightforward UI updates.
- Escalate to high-reasoning model only for:
  - DBGp framing/parsing correctness
  - protocol/state bugs that are hard to reproduce
  - subtle UI-data synchronization issues

## Between-Run Human Handoff Pattern
- After each step, store:
  - what changed
  - exact manual command run
  - pass/fail result
  - one blocker (if any)
- If next change is mechanical (file moves, tiny config edits, running known commands), prefer human execution and only re-engage model for logic-heavy updates.

Last updated: 2026-03-04 by Codex
