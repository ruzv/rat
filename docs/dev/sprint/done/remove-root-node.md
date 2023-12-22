---
id: 784fddef-2b0f-4a9d-8d19-3dc325df19c1
---

# remove-root-node

<rat graph depth=1 />

---

currently all graphs have a root node, that all other nodes stem from. this is
redundant as the root node pretty much never changes and is just an extra perfix
to all node paths.

remove the root node so that graph config requires only one param - path do
directory where nodes are going to be keept
