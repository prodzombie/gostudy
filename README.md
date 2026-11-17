# gostudy

A small command-line tool for Markdown study cards. Terminal review uses Bubble Tea and Bubbles for a scrolling interface, Glamour for Markdown, and golang.org/x/term to detect interactive terminals.

```sh
go run . add -t go -q 'What is a slice?' -a 'A view over an array.'
go run . list
go run . list -t go
go run . show HUID
```

Use `-d DIR` before the command to select a different root directory. By default, cards live under the current directory. `add` prints the new HUID; `list` prints `TAG -> HUID -> question` per line, condensing multiline questions into one line; `show` prints the card file. `help` prints the command summary.

Each card is stored at `cards/TAG/HUID/CARD.md`. Tags are one path component and may contain ASCII letters, digits, `_`, and `-`; a tag must contain at least one non-digit. The file has YAML frontmatter with `id` and `tags` properties, followed by `## Question` and `## Answer` sections. The body can contain Markdown. For example:

```markdown
---
id: "20260930-120000-abcdef123456"
tags:
  - "go"
---

## Question

What is a slice?

## Answer

A view over an array.
```

Options use single-letter names and precede operands. `--` ends option parsing. Usage errors exit with status 2; other errors exit with status 1.

To create a blank card and open it in your editor:

```sh
export EDITOR='vim'
go run . add -e -t go
go run . edit HUID
```

`add -e` also accepts `-q` and `-a` to prefill the card. The editor receives the card's absolute path and uses the current terminal; the CLI waits for it to exit. `$EDITOR` must be set and can contain a command with arguments, such as `code --wait`. An editor error returns status 1; a newly created card remains saved and its ID is printed before the editor opens.

Start a review session for all due cards, or filter by tag:

```sh
go run . review
go run . review -t go
```

Practice all matching cards, including cards that are not due, without saving any progress:

```sh
gostudy review -n
gostudy review -n -t go
```

Practice uses Enter to reveal, then Enter again to advance. Answers are disabled, and card contents and schedules remain unchanged. Enter `q` to stop.

The session shows each question, waits for Enter to reveal the answer, then asks whether you remembered it: enter `r` if you recalled the answer or `f` if you forgot it. Anything else asks again. Enter `q` at either prompt to stop. End of input also stops the session. A card is saved only after `r` or `f`, so stopping preserves completed reviews and leaves the current unanswered card unchanged.

Each card shows its position in the session. After answering, a save confirmation shows the next review interval; in a terminal it stays visible on the next card. The closing summary distinguishes a completed session from an early stop and reports how many cards remain in that session.

When both input and output are terminals (and `TERM` is not `dumb`), review opens a full-screen interface. Progress and controls stay visible while the card scrolls. Enter reveals the answer and scrolls toward its heading; scroll upward to revisit the question. Use ↑/↓ or j/k, Page Up/Down or space, Home/End, or the mouse wheel. Press `?` for additional help. After the answer is revealed, press `r` (recalled) or `f` (forgot) to select, then Enter to save; Backspace clears the selection. Quitting restores the previous terminal screen and prints a session summary.

Markdown uses a compact Glow-inspired palette and a prose width capped at 80 columns. Purple headings and actions, subdued navigation, green save feedback, and syntax-highlighted code distinguish content from controls. The palette adapts when the terminal reports a light or dark background, defaulting to dark if no response arrives. Set `NO_COLOR=1` to disable explicit colors while retaining typographic emphasis. Resizing re-renders the content for the new width. Long code lines may still wrap, and inline code can break at hyphens. Very small terminals display a resize message. Redirected input/output and `TERM=dumb` use the line-based review flow with raw Markdown and no screen control sequences.

Scheduling follows SSP-MMC. Each card has a memory half-life in days and a difficulty from 1 to 18. After t days, the chance of recalling a card with half-life h is modelled as 2^(-t/h): one half-life after a review, recall is a coin toss. A recall lengthens the half-life, and more so the longer the wait actually was, counted in local calendar days since the last review (at least one). A lapse resets the half-life, usually to a much shorter one, and raises the difficulty by 2. The next interval is the whole number of days that minimises the expected total review time until the half-life reaches about a year. gostudy finds these intervals by solving a stochastic shortest path problem once per session, when the first answer is saved; this takes a fraction of a second.

New cards are immediately due. The first review starts a card at difficulty 5, a half-life of about 1.8 days, whatever the answer; this fixed starting difficulty is gostudy's choice. `last_reviewed` stores the full UTC timestamp, but due status uses local calendar days: a card is due from local midnight of the last review's local date plus the interval.

Two behaviours are expected rather than bugs. Intervals get shorter as the half-life nears the goal, so a well-known card may come back after only a few days once or twice before it passes the goal. From a half-life of about 357 days, cards are scheduled for 90% recall, about two months and growing; this rule is gostudy's own.

The memory model and its parameters come from Ye, Su and Cao, "A Stochastic Shortest Path Algorithm for Optimizing Spaced Repetition Scheduling" (KDD 2022), with parameter values from the MIT-licensed [maimemo/SSP-MMC](https://github.com/maimemo/SSP-MMC) repository, Copyright (c) 2022 MaiMemo, Inc. The parameters were fitted on that app's vocabulary learners, not on you, so the intervals are optimal only under that model.

Review progress is stored as four flat frontmatter properties: `half_life` (days), `difficulty`, `interval` (days), and `last_reviewed` (an RFC 3339 timestamp). Cards with properties from the earlier SM-2 scheduler (`ease_factor` or `repetitions`) are treated as new, and those properties are removed when the card is next reviewed. Other properties and the Markdown body are preserved when saving. Cards must have nonempty `## Question` and `## Answer` sections; malformed cards or scheduling properties produce an error. Keep scheduling properties in the generated flat scalar format. Edit card bodies and custom properties freely; avoid editing a card concurrently with a review session.

For development, use `make build` to create `bin/gostudy`, `make check` to run tests and vet, and `make fmt` to format the source. `make run ARGS='review -t go'` runs the CLI with arguments. Run `make help` to list all targets.

Example cards under `examples/cards/` teach the repository's layout, package responsibilities, commands, review scheduling, and development workflow. Explore or study them with:

```sh
make run ARGS='-d examples list'
make demo
```

`make demo` reviews a temporary copy of the examples and removes it when the session ends. Each demo starts fresh and leaves the original examples unchanged. To keep your progress, copy `examples/cards/` into a personal directory and review that directory with `-d DIR`.

Licensed under the [BSD Zero Clause License (0BSD)](LICENSE), which permits use, modification, and redistribution without attribution requirements.
