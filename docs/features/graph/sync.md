---
id: e7d6c7ba-bd42-4eb9-9252-665eac89a78c
---

# sync

<rat graph />

automatically pull changes from a repo where a graph is stored and push new
changeswith a configurable interval.

## config

```yaml
port: 8888
graph:
  name: notes
  path: /graph/path
  sync:
    interval: 1m
    keyPath: /path/to/private/key
```
