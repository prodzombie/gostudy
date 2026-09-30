---
id: "20260930-120002-packages"
tags:
  - "gostudy"
ease_factor: 2.5
interval: 1
repetitions: 0
last_reviewed: "2026-09-30T21:51:18.131277Z"
---

## Question

Which package owns each responsibility in gostudy?

## Answer

`main.go` connects the process arguments and terminal streams to `internal/cli`.

- `internal/cli` parses flags, dispatches commands, and reports errors and exit statuses.
- `internal/card` defines card and tag types and validates tag names.
- `internal/cardstore` locates cards, reads Markdown, and persists review schedules.
- `internal/huid` generates and validates identifiers.
- `internal/review` runs the question, reveal, and grade interaction.
- `internal/schedule` applies SM-2 grades to scheduling state.
- `internal/editor` launches the command configured in `$EDITOR`.

The scheduler takes a schedule, grade, and time and returns an updated schedule. It does not read files or prompt the user.
