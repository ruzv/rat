---
id: c680653d-b0ba-4fb0-9b42-f1eec083e16c
---

# todo

<rat graph />

todo markdown element is a code block (opened by three back ticks and closed by
three back ticks) with language set to `todo`

example (im using thee dots ... instead of three backticks ```)

```
...todo

...
```

## entries

a single todo entries starts with `-` or `x`. `x` indicated that the entries has
been marked as done, and `-` as not done.

following the `-` or `x` seperated by space entry can have any markdown text.
single entry can span multiple lines. the end of an entry is considered to be
the start of another entry or the end of the todo.

example

```
...todo
- `important`

  multiline

x this is done
...
```

## hints

todos can have hints. hints are key value pairs that add additional information
about a todo

available hints are

- `due`
  - the due date of a todo
  - allowed time formats are
    - `02.01.2006`
    - `02.01.2006.`
    - `2.01.2006`
    - `2.01.2006.`
    - `2.01`
    - `2.01.`
- `size`
  - durration - amount of time this todo will take to do
  - format - 1h2m3s
- `priority`
  - positive int - how important is the todo
  - 0 - highest priority
- `tags`
  - list of coma separated strings that classify a todo

example

```
...todo
due=02.01.2026
size=1h
priority=3
tags=hey,there

- do some thing
...
```

additionally a `src` tag is rendered, it indicates the node path this todo is
from

---

```todo
x redo hint parsting and formatting, every thin has two functions parse and
  format. define a map form hintType to struct that contains two functions for
  parsing and formatting
x render hints in a consistent order
x hints `priority`
x hints `tags`

x document in features about hints

- allow comments in todo lists, starts with # or //, allow both
- allow sub tasks in todo lists
  would look like this in MD
  - main task
    - sub task 1
    - sub task 2
- progress bar
 <progress value="32" max="100" ></progress>
```
