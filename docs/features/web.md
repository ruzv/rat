---
id: b8697bf1-e8f2-4e65-b9b3-3a80e4655169
---

# Web-App

<rat graph />

---

Rat frontend web app is built with react and statically compiled to a single
page making it a client side rendered page that requires JS to run.

When rat server is compiled, first the web app is build and then embedded in the
server binary. Bundling all parts of the rat in a single binary.

The web-app provides the following routes:

- `GET` `/view`
  - the root page
- `GET` `/view/{path:.*}`
  - view a node

## Console

Console is the top navigation bar visible when viewing a node (`/view` and
`/view/{path:.*}` endpoints).

It provides the following information:

- Node ID (if clicked, the value is copied to clipboard)
- Node path (if clicked, the value is copied to clipboard)

### Buttons

- First button with the root symbol navigates to the root node.
- The following buttons navigate to parent nodes of current node.
- Search button opens the search modal.
- New node button opens the new node modal.
- Delete node button deletes the current node.

The search and create buttons also have shortcuts:

- `control+k` - search
- `control+shift+k` - create a new node

## kanban

drag and drop
