---
id: d97e4584-2c7c-41ad-a521-e4033af475d5
---

# weight

<rat graph />

---

`weight` field in nodes header can be an integer from 0-inf.

- 0 means - no weight, sort by name.
- weight = 1 means place this node at the very top.
- 2 means place this node after nodes with weight 1.
- and so on

if two or more nodes have the same weight, then they are sorted by name.

nodes with a weight are always placed before nodes without a weight.

weights for new sub nodes can be templated.
