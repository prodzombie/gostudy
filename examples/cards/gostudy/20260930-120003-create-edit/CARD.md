---
id: "20260930-120003-create-edit"
tags:
  - "gostudy"
---

## Question

How do you create a card and open it in your editor?

## Answer

Create a complete card with text supplied as flags:

```sh
make run ARGS='add -t go -q "What is a slice?" -a "A view over an array."'
```

Create a blank template and edit it:

```sh
export EDITOR='vim'
make run ARGS='add -e -t go'
```

`add` prints the new HUID. Reopen the card with `edit HUID`, replacing `HUID` with that printed identifier. The CLI waits for the editor to exit. GUI editors need their wait option, such as `EDITOR='code --wait'`.

Fill both question and answer before reviewing. A card created by `add -e` remains saved even if the editor fails.
