
# ddev-xdebug-tui_v1_poc_spec.md

## PoC Scope

Included features:
- Single debug session
- Line breakpoints
- Step controls (step in / over / out / continue)
- Source view
- Stack view
- Variable inspection
- Ephemeral breakpoints

Excluded features:
- Conditional breakpoints
- Watch expressions
- Multi-session support
- Persistence

## Commands

n → step over
s → step in
o → step out
c → continue
b file.php:45 → add breakpoint
rb file.php:45 → remove breakpoint
q → quit

## Acceptance Criteria

A developer can:

1. Run `ddev debug`
2. Set a breakpoint
3. Trigger a request
4. Step through code
5. Inspect variables
6. Exit debugger cleanly
