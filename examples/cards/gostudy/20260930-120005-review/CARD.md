---
id: "20260930-120005-review"
tags:
  - "gostudy"
ease_factor: 2.6
interval: 1
repetitions: 1
last_reviewed: "2026-09-30T21:53:14.092468Z"
---

## Question

What happens during a review, and how is progress saved?

## Answer

`review` selects cards that have never been reviewed or whose next review is due. `review -t TAG` limits the session to a tag.

```sh
make run ARGS='-d examples review -t gostudy'
```

Read the question, press Enter to reveal the answer, then grade your recall from 0 to 5. Grades 0–2 mean forgotten, 3 hard, 4 good, and 5 easy. `q` at either prompt stops the session; end of input also stops it.

Each valid grade is saved immediately. Quitting preserves completed reviews and leaves the current ungraded card unchanged.

SM-2 starts successful intervals at 1 day, then 6 days, then multiplies the previous interval by the ease factor. Forgetting resets the interval to 1 day. Progress lives in the frontmatter properties `ease_factor`, `interval`, `repetitions`, and `last_reviewed`. Other properties and the Markdown body are preserved.
