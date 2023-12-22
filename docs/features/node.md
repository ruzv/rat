---
id: b00cc1f1-a54c-43d3-866a-e3f820026a72
---

# node

<rat graph />

---

a node is a markdown file in a file tree.

```
├──features/
│  ├──web/
│  ├──tokens/
│  │  ├──todo.md
│  │  ├──kanban.md
│  │  └──graph.md
│  ├──node/
│  │  └──templates.md
│  ├──markdown/
│  │  ├──todo.md
│  │  ├──links.md
│  │  └──graphviz.md
│  ├──api/
│  │  └──move.md
│  ├──web.md
│  ├──tokens.md
│  ├──sync.md
│  ├──search.md
│  ├──node.md
│  ├──markdown.md
│  ├──cache.md
│  └──api.md
└──features.md
```

a node consists of two parts

- [ ](4ba640fd-aa69-49a4-a261-b0a34fc2db81)
- content

every node has a UUID (universally unique identifier) identifier that is
generated at creation and stored in header `id` field.

another not so versatile identifier of a node is its path. node's path is the
filepath to markdown file relative to graphs directory without the `.md`
extension

- filepath `/path/to/graph/path/to/node.md`
- node path `graph/path/to/node`

this is less versatile than the `id` of a node because as you re-structure your
graph node paths change, id's don't.
