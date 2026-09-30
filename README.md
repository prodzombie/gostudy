# gostudy

A small command-line tool for Markdown study cards. It uses only Go's standard library.

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

The session shows each question, waits for Enter to reveal the answer, then asks for a grade from 0 to 5. Grades 0–2 mean forgotten, 3 means hard, 4 good, and 5 easy. Enter `q` at either prompt to stop. End of input also stops the session. A card is saved only after a valid grade, so stopping preserves completed reviews and leaves the current ungraded card unchanged.

New cards are immediately due. The SM-2 scheduler gives successful reviews intervals of 1 day, 6 days, then the previous interval multiplied by the ease factor; a forgotten card restarts at 1 day. The next review is due at the last review's UTC timestamp plus the interval in days.

Review progress is stored as four flat frontmatter properties: `ease_factor`, `interval` (days), `repetitions`, and `last_reviewed` (an RFC 3339 timestamp). Other properties and the Markdown body are preserved when saving. Cards must have nonempty `## Question` and `## Answer` sections; malformed cards or scheduling properties produce an error. Keep scheduling properties in the generated flat scalar format. Edit card bodies and custom properties freely; avoid editing a card concurrently with a review session.

For development, use `make build` to create `bin/gostudy`, `make check` to run tests and vet, and `make fmt` to format the source. `make run ARGS='review -t go'` runs the CLI with arguments. Run `make help` to list all targets.
