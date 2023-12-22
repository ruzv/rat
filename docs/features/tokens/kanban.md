---
id: 0bf9811c-94b6-458a-82f9-95cbc291c63d
---

# kanban

<rat graph />

token `rat kanban` that allows to build kanban boards from nodes. columns of
kanban board are nodes. leaf nodes of column nodes are kanban board cards.

columns can be specified with `columns` argument that is a string of coma
seperated list of node IDs that will be kanban boards columns.

```
rat kanban
  columns="
    id1,
    id2,
    id3
  "
```

in the example a kanban board with three columns is defined. the columns are
displayed in the order they are deffined

cards are populated automatically and can be drag and dropped freely from column
to column.
