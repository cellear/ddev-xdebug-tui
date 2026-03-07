# ddev-xdebug-tui

A small, terminal-native debugger for PHP projects running in **DDEV**.

`ddev-xdebug-tui` lets you debug Drupal, Backdrop, WordPress, and other PHP applications **without installing a full IDE**. It provides a simple terminal interface for stepping through code, inspecting variables, and managing breakpoints.

![ddev-xdebug-tui screenshot](WIREFRAMES/ddev-xdebug-tui-screenshot.svg)

This tool is intentionally minimal.

It focuses on the debugging features developers use most often, while remaining small, readable, and easy to understand.

---

# Why This Exists

DDEV v1.25.1 shipped a feature that caught my eye: an interactive terminal dashboard. You type `ddev` at the command line, and instead of a wall of help text, you get a navigable UI — project management, logs, detail views, all in the terminal.

I tried it immediately, and within a few minutes I had an idea: if DDEV can have present things using a terminal UI, could we have a Xdebug interface that doesn't require any manual configuration?  Using Claude, I was able to take that idea to a working proof of concept in less than 48 hours, without exceeding the limits of my $20/month Pro account.

Xdebug debugging is usually done through large IDE integrations such as:

- PhpStorm
- VS Code
- Vim plugins

These tools are powerful but can feel heavy when you just want to:

- set a breakpoint
- step through code
- inspect variables

`ddev-xdebug-tui` provides a lightweight alternative for developers who prefer **terminal workflows**.

---

# Features

Current PoC capabilities:

- terminal UI debugger
- line breakpoints
- step in / step over / step out / continue
- view call stack
- inspect variables
- view source around current execution point
- works with **DDEV projects**

The debugger attaches to the **first incoming Xdebug session**.

Breakpoints are **ephemeral per run**.

---

# Installation

From your DDEV project directory:

```
ddev add-on get cellear/ddev-xdebug-tui
```

That's it. The binary downloads automatically the first time you run `ddev xdebug-tui`.
No Go installation required.

**Developers:** If you have Go installed and want to build from source, `make install`
builds and installs the binary to `~/go/bin/` and takes precedence over the auto-download.

---

# Usage

From your DDEV project directory:

```
ddev xdebug-tui
```

Xdebug is enabled automatically on start and disabled when you quit.

Then trigger a request in your browser or with `curl`.

The debugger will pause at your first breakpoint (or at the first executable line on entry).

---

# Commands

```
s  step into
n  step over
o  step out
r  run (continue to next breakpoint or end)
q  quit
```

Breakpoint commands:

```
b file.php:45
rb file.php:45
```

---

# Philosophy

This project deliberately avoids becoming a full-featured Xdebug client.

Goals:

- small codebase
- easy to understand
- easy to run
- minimal configuration
- CLI-first workflow

If you need advanced debugging features such as:

- conditional breakpoints
- watch expressions
- multiple concurrent sessions

you should use a full IDE debugger.

---

# Project Status

Early Proof of Concept.

The current goal is a stable terminal debugger for basic stepping and variable inspection.

---

# How This Was Built

This project was built almost entirely by AI agents working under human direction,
across multiple sessions and models. The development used a structured **handoff
protocol** (see `.agent-handoff/`) to maintain continuity across context windows,
a semi-scrum sprint structure with explicit demo checkpoints, and deliberate model
tiering — reserving more capable models for architecture and protocol work, faster
models for mechanical implementation. Each sprint produced a `LEARNINGS/` document
aimed at PHP developers encountering Go concepts for the first time. The full
account is in [DEVELOPMENT_PROCESS.md](DEVELOPMENT_PROCESS.md).

---

# Contributing

Contributions are welcome, but please keep the project philosophy in mind:

- avoid unnecessary complexity
- keep the code understandable
- prefer simple implementations

See:

```
AGENT.md
REPO_BOOTSTRAP.md
```

for implementation guidance.

---

# License

MIT
