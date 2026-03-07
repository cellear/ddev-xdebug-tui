# Handoff — Sprint 5 — 2026-03-07

## Model Attribution

| Role | Model |
|------|-------|
| Sprint planning, all implementation, LEARNINGS, handoff | claude-sonnet-4-6 |

---

## What Was Attempted

Sprint 5 added binary distribution so testers and users can install
`ddev-xdebug-tui` without having Go or the source code. All four stories
implemented in one session.

### S5-1 — Self-installing host command

`commands/host/xdebug-tui` now has a three-tier binary resolution:

1. If `ddev-xdebug-tui` is in PATH (Go developer, `make install`) — use it
2. If `~/.ddev-xdebug-tui/ddev-xdebug-tui` exists (previously auto-installed) — use it
3. Otherwise: detect OS (`uname -s`) and arch (`uname -m`), download from
   `github.com/cellear/ddev-xdebug-tui/releases/latest/download/ddev-xdebug-tui-{os}-{arch}`,
   install to `~/.ddev-xdebug-tui/`, make executable

Unsupported OS or arch exits with a clear error message pointing to the releases
page. The `latest` URL means the script never needs updating when new versions
ship.

### S5-2 — Makefile `dist` target

`make dist` cross-compiles for darwin/amd64, darwin/arm64, linux/amd64,
linux/arm64 using `GOOS`/`GOARCH` env vars in a shell loop. Output goes to
`dist/` (added to `.gitignore`).

### S5-3 — GitHub Actions release workflow

`.github/workflows/release.yml` triggers on `v*` tag push. Builds all four
platform binaries on `ubuntu-latest` and publishes them as assets on a GitHub
Release via `softprops/action-gh-release@v2`. Auto-generates release notes.

### S5-4 — README simplified

Installation section reduced to one command: `ddev add-on get cellear/ddev-xdebug-tui`.
Developer build-from-source instructions moved to a one-liner note.

---

## What Worked

Everything was straightforward. Go cross-compilation is genuinely zero-friction —
the `GOOS`/`GOARCH` approach just works. The `releases/latest/download/` URL
pattern on GitHub is stable and well-documented.

---

## What Didn't Work / Watch Out For

- **`softprops/action-gh-release@v2` requires `permissions: contents: write`.**
  Without this the workflow will fail with a 403 when trying to create the release.
  Already included in the workflow file.

- **The binary has NOT been downloaded and tested yet.** The GitHub Release does
  not exist yet — the human must cut it. Until the release is published, the
  auto-download will 404 and the error message will direct users to the releases
  page.

- **Windows is unsupported.** The script exits with a message. DDEV runs on
  Windows via WSL2, where `uname` will return `Linux` — so WSL2 users will
  actually get the linux binary and it should work fine.

---

## Current State

Code is committed and ready to push. The human needs to:

**To cut the initial v0.4.0 release (one time):**
```bash
make dist
gh release create v0.4.0 dist/ddev-xdebug-tui-* \
  --title "v0.4.0 — Initial release" \
  --notes "First public release. Install via: ddev add-on get cellear/ddev-xdebug-tui"
```

**For all future releases (after pushing code + GitHub Actions):**
```bash
git tag v0.5.0
git push --tags
# GitHub Actions builds and publishes automatically
```

---

## Files Created or Modified

**Created:**
- `.agent-handoff/SPRINTS/sprint-5.md`
- `.agent-handoff/HANDOFF/handoff-2026-03-07-sprint-5-claude-sonnet-4-6.md` — this file
- `.github/workflows/release.yml`
- `LEARNINGS/sprint-5.md`

**Modified:**
- `commands/host/xdebug-tui` — three-tier binary resolution + auto-download
- `Makefile` — added `dist` target, updated `clean` to remove `dist/`
- `.gitignore` — added `bin/` and `dist/`
- `README.md` — simplified installation to single `ddev add-on get` command

---

## References

- `.agent-handoff/AGENT.md` — handoff protocol
- `.agent-handoff/SPRINTS/sprint-5.md` — sprint plan
- `LEARNINGS/sprint-5.md` — Go cross-compilation, GitHub Releases, self-installing scripts

Last updated: 2026-03-07 by claude-sonnet-4-6
