# Handoff: GUI Options Research

**Date:** 2026-03-28  
**Analysis by:** claude-opus-4-6 (Ask mode)  
**Documented by:** claude-sonnet-4-6 (Agent mode — packaged analysis into DOC + handoff, authored commit)  
**Session type:** Research / Documentation only — no code changes

---

## What Was Done

The user asked three questions about potential UI directions for the project:

1. Can panes be different colors in the current TUI?
2. Can the tool be ported to a desktop GUI (like Tk, but more modern)?
3. Can the tool have a pop-up web/HTML or React interface?

These were analyzed in Ask mode, then documented in Agent mode.

---

## Outcome

- **No code was changed.**
- A new reference doc was created: `.agent-handoff/DOC/gui-options.md`

---

## Key Findings

### Colored panes
Trivial. tview/tcell already supports per-widget color via `SetBackgroundColor()`, `SetTextColor()`, `SetBorderColor()`. The status bar already uses this. A ~15-minute change, no new dependencies.

### Desktop GUI
Several Go-native options exist. **Fyne** is the most practical: it has a strong widget set, resizable panes, mouse support, and is cross-platform. The main cost is ~20 MB binary size increase. The backend packages (`dbgclient`, `breakpoints`, `source`) would be unchanged; only `internal/tui` would be replaced.

### Web interface
Two sub-options: **Wails** (native window wrapping a React frontend, single binary, no browser needed) or an **embedded HTTP server + WebSocket** (browser tab, maximum web tooling flexibility). Both keep the Go backend intact.

---

## Current Project State

The project is in a working PoC state as of sprint 5. The TUI is functional. No open bugs were discussed in this session.

---

## Open Questions

- Is the project's minimal-dependency philosophy (see `AGENT.md`) flexible enough to allow Fyne or Wails? These would be significant additions.
- If going the web route, would Wails (native window feel) or browser-tab (simpler architecture) be preferred?
- Are colored panes something the maintainer wants to do soon, or is this just exploratory?

---

## Files Created

- `.agent-handoff/DOC/gui-options.md` — new reference doc, GUI options analysis

## Files Modified

- None

---

## References

- `AGENT.md` — project philosophy and constraints (relevant: "minimal dependencies", "avoid heavy frameworks")
- `.agent-handoff/DOC/gui-options.md` — the primary output of this session
- Prior handoffs: `handoff-2026-03-07-post-release-claude-sonnet-4-6.md` (most recent before this session)
