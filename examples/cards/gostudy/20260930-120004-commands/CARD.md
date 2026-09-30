---
id: "20260930-120004-commands"
tags:
  - "gostudy"
---

## Question

How do you list and show cards, and where do command flags go?

## Answer

`list` prints one line per card in the form `TAG -> HUID -> question`. It condenses multiline questions onto one line and orders cards by tag and HUID. Use `-t TAG` to filter the listing.

```sh
make run ARGS='-d examples list -t gostudy'
```

`show HUID` prints the complete saved Markdown file. Replace `HUID` with an identifier from the listing:

```sh
make run ARGS='-d examples show 20260930-120001-layout'
```

Global `-d DIR` comes before the command. Command options come after the command and before operands. `--` ends option parsing. Usage errors return status 2; other errors return status 1.
