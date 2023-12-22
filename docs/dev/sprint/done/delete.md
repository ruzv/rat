---
id: e9ac4bae-4b6e-4d98-ba2b-e965ebd425bc
---

# delete-in-fe

<rat graph />

allow deleating notes from frontend

create a delete api endpoint.

figure out ui to trigger it. disallow deleating non-leaf nodes.

```sh
curl localhost:8889/graph/nodes/notes/projects/programming/rat/dev/test \
  -X DELETE
```

```todo
x delete button in fe
x delete api call on click, redirect to parent
x confirm modal
```
