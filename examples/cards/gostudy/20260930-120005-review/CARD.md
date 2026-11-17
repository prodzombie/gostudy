---
id: "20260930-120005-review"
tags:
  - "gostudy"
---

## Question

What happens during a review, and how is progress saved?

## Answer

`review` selects cards that have never been reviewed or whose next review is due. `review -t TAG` limits the session to a tag.

```sh
make run ARGS='-d examples review -t gostudy'
```

Read the question, press Enter to reveal the answer, then answer `r` if you recalled it or `f` if you forgot it. `q` at either prompt stops the session; end of input also stops it.

Each answer is saved immediately. Quitting preserves completed reviews and leaves the current unanswered card unchanged.

Each card has a memory half-life in days and a difficulty from 1 to 18. A recall lengthens the half-life, more so after a longer wait; forgetting resets it, usually to a much shorter one, and makes the card harder. The SSP-MMC scheduler then picks the interval that minimises expected total review time under its memory model. A new card starts at difficulty 5, about a 1.8-day half-life.

Progress lives in the frontmatter properties `half_life`, `difficulty`, `interval`, and `last_reviewed`. Cards with old SM-2 properties start again as new. Other properties and the Markdown body are preserved.
