---
id: "20260930-120006-development"
tags:
  - "gostudy"
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

The terminal interface uses Bubble Tea, Bubbles, Lip Gloss and Glamour; the rest uses Go's standard library. Tests cover the card workflow, editor invocation, Markdown preservation, scheduling, due-card selection, and interactive answers.

A typical development loop is: change one behavior, run `make fmt`, then run `make check build`.
