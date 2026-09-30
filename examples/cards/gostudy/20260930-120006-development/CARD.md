---
id: "20260930-120006-development"
tags:
  - "gostudy"
ease_factor: 2.5
interval: 1
repetitions: 0
last_reviewed: "2026-09-30T21:53:23.188401Z"
---

## Question

Which Makefile targets help you work on this repository?

## Answer

Run `make help` to see the available targets.

- `make build` builds the executable at `bin/gostudy`.
- `make run ARGS='help'` runs the CLI with arguments.
- `make test` runs all Go tests.
- `make vet` checks for common Go mistakes.
- `make check` runs tests and vet.
- `make fmt` formats the Go source.
- `make clean` removes the built CLI.

The application uses only Go's standard library. Tests cover the card workflow, editor invocation, Markdown preservation, due-card selection, and interactive grading.

A typical development loop is: change one behavior, run `make fmt`, then run `make check build`.
