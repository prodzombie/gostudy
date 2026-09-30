---
id: "20260930-120001-layout"
tags:
  - "gostudy"
---

## Question

Where does gostudy store a card, and what belongs in its frontmatter?

## Answer

A card lives at `cards/{TAG}/{HUID}/CARD.md` relative to the selected root. Use `-d DIR` before the command to choose that root; the default is the current directory.

The Markdown file begins with YAML frontmatter containing `id` and a `tags` list. Its body has `## Question` and `## Answer` sections. The directory layout supplies the card's identity and tag when the CLI reads it.

These examples use `examples` as their root:

```sh
make run ARGS='-d examples list'
```

A HUID contains a UTC timestamp and an optional suffix. New cards get a random suffix to reduce collisions.
