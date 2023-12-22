---
id: d26fc131-e854-4a8e-8a90-a3684cb9a1c0
---

# cache

<rat  graph />

## path cache

a wrapper for filesystem graph provider that stores in memory nodes paths and id

## full cache

store nodes info fully in memory

- store nodes as reads / writes happen. (don't load the whole graph at once)
- ensure up to date'ness with node data on filesystem
  - `os.stat` checkeck last modified time
