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

## 2. GitHub Releases: two things with the same name

It's easy to confuse "a GitHub Release" with "the git repository." They are
completely separate:

**The git repository** is what you've been working with all along — commits,
branches, files, history. Binaries are gitignored and never go here.

**A GitHub Release** is a separate storage area attached to a version tag. It
has its own page (`github.com/{owner}/{repo}/releases`) and can hold binary
file "assets" — files that are uploaded to GitHub but are NOT part of the git
history at all. Think of it like an email attachment: the email (git repo) and
the attachment (binary) travel together but are stored differently.

When users download software from a GitHub Release, they are downloading those
attached assets — not cloning the repository.

The download URL follows a predictable pattern:

```
https://github.com/{owner}/{repo}/releases/download/{tag}/{filename}
https://github.com/{owner}/{repo}/releases/latest/download/{filename}
```

The `latest` variant always resolves to the most recent non-prerelease release.
This is what the self-install script uses — the URL never needs updating when
new versions ship.

---

## 3. GitHub Actions: yes, GitHub compiles your code in the cloud

This is the part that surprises many developers coming from interpreted languages
like PHP. GitHub offers free cloud computing time for open-source projects, and
you can use it to run arbitrary code — including compiling a Go binary.

Here's what actually happens when you push a version tag:

1. GitHub detects the tag and looks for matching workflow files in `.github/workflows/`
2. GitHub spins up a fresh virtual machine (a Linux server in Microsoft's cloud)
3. That VM clones your repository, installs Go, and runs your build commands
4. The compiled binaries exist on that cloud VM
5. The workflow uploads them as assets to a new GitHub Release
6. The cloud VM is discarded — it existed for maybe 2 minutes

The workflow file is just a recipe. It runs on GitHub's infrastructure, not
yours. You never see the VM or interact with it — you push a tag, wait a minute,
and a Release appears with four downloadable binaries attached.

For PHP developers: this is roughly equivalent to a service that takes your
Drupal codebase, runs `composer install`, zips the result, and posts the zip
somewhere for download — except for compiled binaries instead of PHP packages.

The workflow that does this lives in `.github/workflows/release.yml`. The key
sections:

```yaml
on:
  push:
    tags:
      - 'v*'       # run this workflow whenever a v* tag is pushed
```

```yaml
- uses: actions/setup-go@v5    # install Go on the cloud VM
  with:
    go-version-file: go.mod    # use whatever version our project requires
```

```yaml
- name: Build binaries for all platforms
  run: |
    GOOS=darwin GOARCH=arm64 go build -o dist/ddev-xdebug-tui-darwin-arm64 ...
    GOOS=linux  GOARCH=amd64 go build -o dist/ddev-xdebug-tui-linux-amd64  ...
    # etc.
```

```yaml
- uses: softprops/action-gh-release@v2   # upload dist/* to the GitHub Release
  with:
    files: dist/*
```

The `permissions: contents: write` grant is required — without it the workflow
cannot create releases or upload assets.

After this workflow is in place, cutting a new release is two commands:

```bash
git tag v0.5.0
git push --tags
```

Everything else happens in the cloud.

---

## 4. Self-installing shell scripts

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

## 5. `uname` for OS and architecture detection

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

Compiled binaries should never be committed to git — they're large, platform-
specific, and opaque (you can't diff them or review them meaningfully). Add to
`.gitignore`:

```
bin/
dist/
```

This means `make dist` creates binaries on your local machine that git ignores
completely. They exist temporarily so you can upload them to a GitHub Release
with `gh release create`. Once uploaded, they live as Release assets — not in
git — and you can delete the local `dist/` folder.

The mental model: **git stores your source, GitHub Releases store your
compiled output.** They are different systems that happen to live on the same
website.

Last updated: 2026-03-07 by claude-sonnet-4-6
