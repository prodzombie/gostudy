---
id: "20260930-120002-packages"
tags:
  - "gostudy"
---

## Question

Which package owns each responsibility in gostudy?

## Answer

`main.go` connects the process arguments and terminal streams to `internal/cli`.

- `internal/cli` parses flags, dispatches commands, and reports errors and exit statuses.
- `internal/card` defines card and tag types and validates tag names.
- `internal/cardstore` locates cards, reads Markdown, and persists review schedules.
- `internal/huid` generates and validates identifiers.
- `internal/review` runs the question, reveal, and recalled-or-forgot interaction.
- `internal/schedule` models memory and applies SSP-MMC scheduling to review answers.
- `internal/calendar` converts times into local calendar days.
- `internal/editor` launches the command configured in `$EDITOR`.

The scheduler takes a schedule, whether the card was recalled, and a time, and returns an updated schedule. It does not read files or prompt the user.
