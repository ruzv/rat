---
id: ece4682b-f811-4f5f-9e1e-e72754d1eb2b
---

# todo

<rat  graph />

## todo token

- `rat todo` in `<>` brackets (cant do here, cause it would get parsed)
- arguments
  - `sources`
    - a list of coma seperated node ids whose child nodes will be used to search
      for matching todos, the specified node id is itself also included in the
      search space
    - add a prefix `-` to an id to exclude this node an all of its child nodes
      from the search
    - example `sources="id1,-id2,id3"`
  - `filter_has`
    - a list of coma separated hint keys, to filter todos that have a particular
      hint
    - available hints
      - `due`
      - `size`
      - `priority`
      - `tags`
    - add a prefix `!` to flip the search condition - does not have a hint
    - example `filter_has="!due,size"`
  - `filter_value`
    - a string that contains list of coma seperated expresions, that will filter
      todos with hints for which the specified expresions are true
    - `due=1` `due<4` `due>1`
      - due expresions are evaluated by converting the due date into days - how
        soon the todo will become due, hence the expresion `due<4` would mean,
        filter all todos whose due date is in less than 4 days.
    - `size=5h` `size<5h` `size>5h`
    - `priority=1` `priority<1` `priority>2`
    - `tags=fun`
      - `=` only equals operator is processed as a check if the todos tags
        contains the specified tag after equal sign. example `tags=important`
      - `>` `<` are ignores
  - `include`
    - by default all done todo entries and all done todos (todos whose entries
      are all done) are not rendered
    - `include=done` - render todos that have been compleated
    - `include=done_entries` - render todo entries that have been compleated
    - example `include=done,done_entries`
  - `sort`
    - list of coma seperated hints
    - sort the found todos in some specified order
    - available sort terms
      - `due` `-due`
      - `size` `-size`
      - `priority` `-priority`
    - by default sort order is increasing, adding prefix `-` reverses the order
    - example `sort="due,-size,priority"`

---

```todo
x handle emptie todo list for when trannsfroming token (i think its done, test)
x allow specifying the depth
x filter_has
x token "include done entries" arg
x todo token sort arg
x remake include_done, as a `include` filter, that would have options like
  `done`, `done_entries`
x filter by hint value
  due>3
  due~week  -  7days
  due_in=X - due in X days - due < today + X
  was_due=X -  due > today + X
  complex numerical
  priority=1,3
  simple numerical - greater, less, equal
  size<1h
  simple numerical
  tags=m,k,j
  contains
x document in features about args, make sure to mention default values
x the ability to list multiple source nodes
  soruces=-id1,id2
  prefix - would exclude all todo form that source
  incompatible with parent arg
x better errors. display path to node where error happened
```
