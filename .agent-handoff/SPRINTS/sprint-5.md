# Sprint 5 — Release & Distribution

**Goal:** Make `ddev-xdebug-tui` installable by testers and users who don't have
Go — no source checkout, no compilation required.

**Demo:** Tester with no Go installed runs `ddev add-on get cellear/ddev-xdebug-tui`
in a DDEV project, then `ddev xdebug-tui`. Binary auto-downloads and debugger
launches.

---

## Stories

### S5-1 · Self-installing host command · [done]

**Owner:** Sonnet
**Scope:** Update `commands/host/xdebug-tui` to detect whether the binary is
already installed (in PATH or in `~/.ddev-xdebug-tui/`), and if not, download
the correct binary from the latest GitHub Release.

**Acceptance criteria:**
- If `ddev-xdebug-tui` is in PATH (e.g. from `make install`), use it directly
- If binary previously auto-installed to `~/.ddev-xdebug-tui/`, use it directly
- On first run with no binary: detect OS (darwin/linux) and arch (amd64/arm64),
  download from `github.com/cellear/ddev-xdebug-tui/releases/latest/download/`,
  install to `~/.ddev-xdebug-tui/ddev-xdebug-tui`, make executable
- Unsupported OS/arch prints a helpful error and exits cleanly

### S5-2 · Makefile dist target · [done]

**Owner:** Sonnet
**Scope:** Add a `make dist` target that cross-compiles binaries for all supported
platforms: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64.

**Acceptance criteria:**
- `make dist` produces four binaries in `dist/`
- Naming convention: `ddev-xdebug-tui-{os}-{arch}`
- `dist/` added to `.gitignore`

### S5-3 · GitHub Actions release workflow · [done]

**Owner:** Sonnet
**Scope:** Add `.github/workflows/release.yml`. On push of a `v*` tag, build
binaries for all platforms and attach them to a GitHub Release automatically.

**Acceptance criteria:**
- Triggers on `v*` tag push
- Builds darwin/amd64, darwin/arm64, linux/amd64, linux/arm64
- Attaches all four binaries to the release
- Uses `actions/setup-go` with version from `go.mod`

### S5-4 · README — simplified installation · [done]

**Owner:** Sonnet
**Scope:** Update installation section. Users no longer need Go or a source
checkout. Single `ddev add-on get` command is sufficient.

---

## Demo · Tester Flow

```
# In a DDEV project directory:
ddev add-on get cellear/ddev-xdebug-tui
ddev xdebug-tui
# → "ddev-xdebug-tui: installing binary..."
# → "Downloading from https://github.com/..."
# → "Installed to ~/.ddev-xdebug-tui/ddev-xdebug-tui"
# → TUI launches
```

**Demo status: [ ] PENDING — requires GitHub Release to be cut by human**

Human steps to cut release (after pushing this sprint):
```bash
make dist
gh release create v0.4.0 dist/ddev-xdebug-tui-* \
  --title "v0.4.0 — Initial release" \
  --notes "First public release. Install via: ddev add-on get cellear/ddev-xdebug-tui"
```

---

## Notes

- `dist/` must be in `.gitignore` — binaries should not be committed
- The `releases/latest/download/` URL redirects correctly for the most recent
  non-prerelease; no version pinning needed in the script
- Future releases: `git tag v0.5.0 && git push --tags` triggers the workflow
