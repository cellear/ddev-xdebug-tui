# LEARNINGS — Sprint 5

> Concepts encountered during Sprint 5 (Release & Distribution).
> Written for PHP developers learning Go and DevOps tooling.

---

## 1. Go cross-compilation with `GOOS` and `GOARCH`

Go can compile binaries for any supported platform from any machine. You don't
need a Mac to build a macOS binary or a Linux box to build a Linux binary. Two
environment variables control the target:

- `GOOS` — operating system (`darwin`, `linux`, `windows`)
- `GOARCH` — CPU architecture (`amd64` for Intel/AMD 64-bit, `arm64` for Apple
  Silicon and modern ARM)

```bash
GOOS=darwin GOARCH=arm64 go build -o dist/ddev-xdebug-tui-darwin-arm64 ./cmd/ddev-xdebug-tui/
```

This single command on a Linux CI runner produces a binary that runs on an
Apple Silicon Mac. Go's standard library has no C dependencies by default, so
cross-compilation just works.

In PHP terms: there's no equivalent, because PHP is interpreted. This is one of
the advantages of compiled languages for distribution — you ship a single static
binary with no runtime requirement.

---

## 2. GitHub Releases as a binary distribution channel

GitHub Releases are versioned snapshots of a repository, optionally with
attached binary files called "assets." The download URL follows a predictable
pattern:

```
https://github.com/{owner}/{repo}/releases/download/{tag}/{filename}
https://github.com/{owner}/{repo}/releases/latest/download/{filename}
```

The `latest` variant always resolves to the most recent non-prerelease release.
This is useful in auto-install scripts: you never need to update the URL when
you ship a new version.

---

## 3. Self-installing shell scripts

The host command script now installs the binary on first run. The pattern:

```bash
if command -v ddev-xdebug-tui &>/dev/null; then
  BINARY_CMD="ddev-xdebug-tui"          # already in PATH
elif [ -x "$BINARY_PATH" ]; then
  BINARY_CMD="$BINARY_PATH"             # previously auto-installed
else
  # download and install
  curl -fsSL "$URL" -o "$BINARY_PATH"
  chmod +x "$BINARY_PATH"
  BINARY_CMD="$BINARY_PATH"
fi
```

Key shell idioms:
- `command -v foo` — returns the path of `foo` if it's in PATH, exits non-zero
  if not. Prefer over `which` (more portable).
- `[ -x "$path" ]` — true if the file exists and is executable.
- `curl -fsSL` — `-f` fail on HTTP errors, `-s` silent, `-S` show errors,
  `-L` follow redirects. This combination is the standard for scripted downloads.
- `chmod +x` — makes a downloaded file executable (downloads don't preserve
  execute permissions).

---

## 4. `uname` for OS and architecture detection

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')   # "Darwin" → "darwin"
ARCH=$(uname -m)                               # "x86_64" or "arm64"
```

`uname -s` returns the OS name (`Darwin` on macOS, `Linux` on Linux).
`uname -m` returns the machine hardware name (`x86_64`, `arm64`, `aarch64`).

Note: Linux on ARM reports `aarch64` while macOS reports `arm64` for the same
architecture. The `case` statement normalises these:

```bash
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
esac
```

---

## 5. GitHub Actions for Go releases

A GitHub Actions workflow is a YAML file in `.github/workflows/` that runs on
trigger events. For a release workflow:

```yaml
on:
  push:
    tags:
      - 'v*'       # triggers on any tag starting with "v"
```

The workflow uses two community actions:
- `actions/setup-go@v5` — installs Go, version read from `go.mod`
- `softprops/action-gh-release@v2` — creates the GitHub Release and uploads assets

The `permissions: contents: write` grant is required for the workflow to create
releases and upload files to the repository.

After this is in place, cutting a new release is:

```bash
git tag v0.5.0
git push --tags
```

GitHub Actions does the rest — builds four binaries and publishes the release.

---

## 6. Makefile for multi-platform builds

The `dist` target uses a shell loop over a Make variable:

```makefile
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

dist:
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		GOOS=$$os GOARCH=$$arch go build -o dist/ddev-xdebug-tui-$$os-$$arch \
		  ./cmd/ddev-xdebug-tui/; \
	done
```

`$${platform%/*}` strips everything from the last `/` onward (gets `darwin` from
`darwin/amd64`). `$${platform#*/}` strips everything up to the first `/` (gets
`amd64`). The double `$$` escapes the shell variable inside Make's `$(...)`.

---

## 7. Keeping binaries out of git

Compiled binaries should never be committed — they're large, opaque, and
platform-specific. Add to `.gitignore`:

```
bin/
dist/
```

The GitHub Release is the right place for binaries; the git repository is for
source code and documentation.

Last updated: 2026-03-07 by claude-sonnet-4-6
