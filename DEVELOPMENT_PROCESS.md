# Development Process

`ddev-xdebug-tui` was built almost entirely by AI agents, working under human
direction, across multiple sessions and multiple models. This document describes
how that worked — not as an abstract methodology, but as a concrete account of
what we actually did and why.

---

## The Core Problem: Agents Forget

Every AI coding session starts fresh. The agent has no memory of last session's
decisions, bugs, or dead ends. If you don't solve this, you spend half of every
session re-explaining what already happened.

Our solution was a **handoff protocol** — a structured directory of documents
that agents write at the end of each session and read at the start of the next.

```
.agent-handoff/
  AGENT.md          # standing instructions for every agent
  DOC/              # persistent architecture docs
  SPRINTS/          # sprint-by-sprint plans and status
  HANDOFF/          # per-session logs written by agents for agents
```

Every handoff document follows the same format: what was attempted, what worked,
what didn't, current state, open questions, files changed. Agents are instructed
to read these before doing anything else. In practice this meant each session
started with genuine continuity — the agent knew about the iso-8859-1 encoding
bug fixed two sessions ago, the DDEV path conventions we'd established, the
naming decisions we'd made.

The key insight: **handoffs are written for agents, not humans.** They're dense,
technical, and free of pleasantries. They cover the things a future agent needs
to pick up mid-stream without re-deriving context from scratch.

See `.agent-handoff/AGENT.md` for the full protocol and the `HANDOFF/` directory
for every session log from this project.

---

## Semi-Scrum With AI Roles

We ran the project in four sprints, each with an explicit plan document, user
stories with acceptance criteria, and demo checkpoints.

A sprint plan looks like this:

```
### S3-2 · Source panel · [done]

**Owner:** Sonnet
**Scope:** MapPath, ContainerPath, Format; update TUI to call refreshSource
on each step.

**Acceptance criteria:**
- Source panel shows current file with correct line highlighted
- Path mapping works for files in project root and subdirectories
```

The `[done]` tag gets filled in when the story is complete. Demo checkpoints
are explicit milestones — we didn't proceed to the next story until the demo
passed. This prevented the common failure mode of "everything is half-done."

Sprint planning was done by the human and Sonnet together at the start of each
session. Implementation was handed off to cheaper models for mechanical work,
then reviewed by the human via screenshot. The sprint doc became the shared
source of truth that both human and agent could refer to.

---

## Model Tiering

Not all tasks need the same model. We used three tiers:

**High-reasoning model (Sonnet):** Architecture decisions, DBGp protocol
implementation, tricky bug diagnosis, sprint planning, handoff writing. These
are tasks where the wrong decision is expensive — you want the model that thinks
carefully rather than the model that responds quickly.

**Faster model (Haiku):** Mechanical implementation once the design was clear.
"Given this struct definition and these acceptance criteria, implement the
method." Haiku wrote the TCP listener, the breakpoint store, the step command
handlers, and the variables/stack panels. These are tasks where the design is
already settled and the work is translating a spec into code.

**Human:** Running the code, taking screenshots, operating the browser, making
GitHub decisions, pushing to the remote. The human also served as the feedback
loop that agents can't provide for each other — "the highlighting is invisible
on dark terminals" is the kind of observation that only comes from actually
running the thing.

The cost difference between models is significant. By reserving the expensive
model for work that genuinely benefits from it, we kept the project economical
without sacrificing quality where it mattered.

---

## Human-in-the-Loop

One of the persistent failure modes in AI-assisted development is the agent
doing things that only the human should do — pushing to production, modifying
permissions, making irreversible decisions. We avoided this by being explicit
about role boundaries.

Things only the human did:
- Run `go build` and test the binary
- Operate the browser to trigger Xdebug sessions
- Take and provide screenshots
- `git push` to the remote
- Make the repository public
- Make judgment calls on UX ("should we wrap source lines?")

Things agents did:
- Write and modify all Go source files
- Write all documentation
- Manage git commits (but not push)
- Debug protocol issues
- Make low-stakes implementation decisions within established constraints

The screenshot feedback loop was particularly effective. After each demo
checkpoint, the human took a screenshot and shared it. The agent could see
exactly what the user saw — a yellow-highlighted current line, base64-encoded
variable values, an empty stack panel — and respond directly to observed
behavior rather than speculated behavior.

---

## LEARNINGS Docs

Each sprint produced a `LEARNINGS/sprint-N.md` — a document written by the
agent at the end of the sprint, aimed at PHP developers learning Go.

This served two purposes. First, it captured hard-won knowledge (Xdebug
encodes strings as base64; `sync.Once` is the right tool for idempotent channel
close; `exec` in a shell script prevents cleanup code from running) in a form
that survives across sessions. Second, it made the project useful to developers
who might read the codebase later — not just "here is the code" but "here is
what we learned building it."

Writing LEARNINGS at the end of each sprint also functioned as a forcing
function for the agent: articulating what was learned required understanding
it well enough to explain it.

---

## What This Process Is Good For

This approach worked well for this project because:

- The scope was well-defined and bounded from the start
- The technology stack (Go, tview, DBGp) was established rather than experimental
- The human could verify progress incrementally via screenshots and a running binary
- Sessions could be isolated — each sprint was a coherent unit of work

It would work less well for projects where requirements are genuinely unknown,
where the human can't verify agent output easily, or where the codebase is large
enough that handoff documents can't capture the necessary context.

---

## The Numbers

- 4 sprints over 1 day
- 4 handoff documents written by agents, 2 by other tools
- ~1,400 lines of Go across 6 packages
- 4 LEARNINGS docs covering 30+ Go concepts
- Every line of code reviewed by the human before the repo went public

The repository itself is a record of this process. The `.agent-handoff/`
directory is intentionally public — not because it's polished, but because the
raw session logs are more honest and more useful than a cleaned-up retrospective.

---

## Reusing This Approach

The handoff protocol is not specific to this project. The `.agent-handoff/`
directory structure, the `AGENT.md` standing instructions, and the sprint
document format can be adopted for any project where you want continuity across
AI sessions.

The minimum viable version is just two things: a standing instructions file
that tells every new agent how to orient itself, and a log of what happened
last session. Everything else is elaboration on that core idea.
