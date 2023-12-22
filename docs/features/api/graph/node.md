---
id: bc88dca8-a268-4daf-9186-09f75bdf7e12
---

# Node

<rat graph depth=1 />

---

## Read - GET `/graph/node/{path:.*}`

Read a node by its path.

## Create - POST `/graph/node/{path:.*}`

Create a new node.

## Delete - DELETE `/graph/node/{path:.*}`

Deletes an existing node and all of its sub nodes.

```sh
curl localhost:8888/graph/node/{nodePath} \
  -X DELETE
```
