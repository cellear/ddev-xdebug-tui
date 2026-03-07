# Project Overview

title: XDebug Terminal User Interface DDEV Add-on
description: Terminal-native PHP/Xdebug debugger for DDEV — step through code and inspect variables without an IDE
status: complete
notable:
  - lightweight alternative to PhpStorm/VS Code for developers who prefer terminal workflows
  - built in 4 sprints over 1 day: ~1,400 lines of Go across 6 packages
  - used model tiering throughout: Sonnet for architecture/protocol work, Haiku for mechanical implementation
  - semi-scrum sprint structure with explicit demo checkpoints — nothing advanced until current story passed
  - screenshot feedback loop: human ran the binary and shared screenshots; agents responded to observed behavior
  - each sprint produced a LEARNINGS doc aimed at PHP developers encountering Go concepts
  - handoff protocol kept continuity across sessions and models; .agent-handoff/ is public in the repo
  - methodology is explicitly documented in DEVELOPMENT_PROCESS.md and reusable for other projects
blog_candidate: yes — the development process is the story, not just the tool
book_candidate: yes — strong exemplar for the handoff methodology; numbers (4 sprints, 1 day) are concrete
drupalcon_comic: no
