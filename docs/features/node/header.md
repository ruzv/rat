---
id: 4ba640fd-aa69-49a4-a261-b0a34fc2db81
---

# Header

<rat graph />

---

Node header is a YAML header in nodes markdown file that contains metadata about
the node.

Example:

```md
---
id: uuid
weight: 10
name: "🎥 Movies"
template:
  name: "{{ .RawName }}"
---

# Regular content of markdown file
```

A header has the following structure:

```yaml
id: uuid # a uuid string
weight: 10 # a unsigned int (0 and up)
name: "🎥 Movies" # a string
template:
  name: "{{ .RawName }}" # string, that can contain template fields
  weight: 0 # a string, can contain template fields, or unsigned int
  content: | # a string, can contain template fields
    # {{ .Name }}

    <rat graph />

    ---
```

All fields except `id` are optional.

- `id` - unique node identifier, auto-generated and shouldn't really be changed
  ever.
- `name` - a more human-readable name of the node. The last segment of nodes
  path is considered its name. The path segments have restrictions on what
  symbols they can contain. The `name` field by passes those restrictions
  allowing any symbols to be rendered as the nodes name.
- `weight` - determines the order of nodes in kanban columns, graph lists
- `template` - specifies how new sub nodes should be created
