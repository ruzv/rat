---
id: 04d046bb-7f17-4929-a8b8-a46ab85889fe
---

# Graph

<rat graph />

---

Graph is a file system tree of markdown files and directories.

A graph is configured by its provider.

```yaml
services:
  provider:
    dir: /path/to/graph/dir
    enablePathCache: true
    root:
      content: hello
      template:
        name: "{{ .RawName }}"
        weight: 0
        content: |
          # {{ .Name }}

          <rat graph depth=1 />

          ---
```

## File system provider

A graph provider that reads and writes node (markdown file) data to the file
system.

The location where Rat file system provider will write node data is configured
by `services.provider.dir` field in the YAML config file, provided to Rat server
process at start.

## Path cache provider

Path cache provide can be used to improve Rat server performance. It helps other
providers resolve node paths and IDs in a more efficient way, by caching them.

It's enabled by default and can be disabled by setting
`services.provider.enablePathCache` to `false`.

## Root node

Every graph has a special node - the root node, from which all other nodes are
connected.

The root node has path - `""` (empty string).

The root node has ID - `00000000-0000-0000-0000-000000000000` (empty UUID).

The root node just like other nodes has content and template fields. These can
be set in the config `services.provider.root.content` and
`services.provider.root.template` (template is an object with structure, just
like for all other nodes) fields.
