# gostudy

A small command-line tool for Markdown study cards. It uses only Go's standard library.

```sh
go run . add -t go -q 'What is a slice?' -a 'A view over an array.'
go run . list
go run . list -t go
go run . show HUID
```

Use `-d DIR` before the command to select a different root directory. By default, cards live under the current directory. `add` prints the new HUID; `list` prints one HUID and tag per line, separated by a tab; `show` prints the card file. `help` prints the command summary.

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
