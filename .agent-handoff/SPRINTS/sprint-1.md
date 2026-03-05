# Sprint 1: Environment Setup + Project Scaffold

**Sprint Goal:** By the end of this sprint, a developer can type `ddev debug` in a DDEV PHP project and see the ddev-xdebug-tui terminal UI launch and display "waiting for Xdebug connection."

---

## Stories

---

### S1-1: Set Up PHP Test Project with DDEV
**Status:** [done]
**Owner:** Human

**Description:**
Create a minimal PHP project managed by DDEV to serve as the test harness for all future sprints. This does not need to be Drupal — a simple multi-file PHP app is sufficient for PoC testing and much faster to set up.

**Acceptance Criteria:**
- A directory exists with a minimal PHP project (e.g., `index.php` calling functions across 2-3 files)
- `ddev config` has been run; project type is `php`
- `ddev start` succeeds
- `ddev xdebug on` succeeds
- Visiting the site in a browser returns a PHP response
- Xdebug is confirmed enabled: `ddev php -r "echo phpinfo();" | grep -i xdebug`

**Notes:**
Keep the PHP source simple. A 2-3 file app with a function call chain gives enough stack depth to test stepping and variable inspection without framework overhead. A flat PHP project (no Composer, no framework) is preferred.

---

### S1-2: Initialize Go Repository Scaffold
**Status:** [done]
**Owner:** Human or Haiku

**Description:**
Create the Go module and directory structure as defined in REPO_BOOTSTRAP.md. The project should compile (even with empty stubs) before moving to the next story.

**Acceptance Criteria:**
- `go.mod` exists with module path `github.com/cellear/ddev-xdebug-tui`
- Dependencies added: `github.com/rivo/tview`, `github.com/gdamore/tcell/v2`
- Directory structure exists:
  - `cmd/ddev-xdebug-tui/`
  - `internal/dbgclient/`
  - `internal/session/`
  - `internal/breakpoints/`
  - `internal/source/`
  - `internal/tui/`
- Each `internal/` package has a minimal stub file so `go build ./...` succeeds

---

### S1-3: Minimal Runnable TUI Shell
**Status:** [done]
**Owner:** Haiku

**Description:**
Implement a minimal `main.go` that starts the tview application and renders a waiting screen. No debugger logic. The TUI shell established here will be the foundation that later sprints wire data into — so the panel layout should match the final design even if panels are empty.

**Acceptance Criteria:**
- `go run ./cmd/ddev-xdebug-tui` starts without error
- Full-screen terminal UI renders with the intended split-pane layout (Stack | Source / Variables | Breakpoints)
- A status line displays: "ddev-xdebug-tui — waiting for Xdebug connection"
- `q` exits cleanly
- Code is small and readable (target: under 60 lines in main.go)

**Notes:**
Scaffold the full panel layout now rather than a placeholder box. This avoids a layout rewrite in Sprint 4 when panels get populated.

---

### S1-4: DDEV Add-on Stub
**Status:** [done]
**Owner:** Human or Haiku

**Description:**
Create the minimal DDEV add-on structure so that `ddev debug` is a recognized command that launches the binary. The binary only needs to show the waiting screen from S1-3.

**Acceptance Criteria:**
- Add-on directory structure is present in the repo (see Notes)
- Running `ddev get .` from the repo root installs the add-on into the DDEV test project
- `ddev debug` is a recognized command in the DDEV test project after installation
- Running `ddev debug` launches the TUI shell from S1-3
- Developer can confirm the reinstall workflow: rebuild binary → re-run `ddev get .` → test (no full uninstall needed between runs)

**Notes:**
DDEV add-ons require a specific directory structure. At minimum:
- `.ddev/commands/host/debug` — a shell script that invokes the binary
- `install.yaml` — declares files to copy on `ddev get`

During development, `ddev get /path/to/ddev-xdebug-tui` (absolute path) or `ddev get .` from the repo root reinstalls without requiring uninstall first.

---

## Sprint Review Demo Checklist

At the sprint review, the stakeholder should be able to observe:

1. Developer runs `ddev get .` from the ddev-xdebug-tui repo root — installs cleanly
2. Developer navigates to the PHP test project
3. Developer runs `ddev debug`
4. TUI launches and shows the split-pane layout with "waiting for Xdebug connection"
5. Developer presses `q` — exits cleanly

---

## Decisions Made

- **Module path:** `github.com/cellear/ddev-xdebug-tui`
- **PHP test project location:** `testdata/php-test-project/` inside this repo
- **PHP source:** Simple flat PHP (no framework) — `index.php` → `lib/math.php` + `lib/greeter.php`

---

## Deferred to Later Sprints

- Path mapping between DDEV container paths and host filesystem (critical for Sprint 3 source loading — flag it then)
- TCP listener and Xdebug connection (Sprint 2)
- DBGp parsing (Sprint 3)

---

Last updated: 2026-03-05 by Claude Haiku
