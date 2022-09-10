# notes server

custom server for:

- serving existing notes
  - this needs hosting
- managing notes
  - create
  - read
  - update
  - delete

## features

- the site is structured like a file tree.
  - file - markdown file containing the content of a single page.
  - directory - bundles files and other directories together.
- markdown files are converted to html rendered as static sites
- links in markdown files to other sites. are like file paths.
- selections

## tools

- go gin https://github.com/gin-gonic/gin
- markdown to html https://github.com/gomarkdown/markdown

## serving

[private server](../private-server/_index.md)

# github

https://github.com/ruzv/rat

# plan

- built as a graph
- every page is a node

  - reperesented as a pair of file and a dir
  - the dir can contain other nodes file and dir pair these nodes are child
    nodes

# tasks

- bugs - [x] rat regex works only if each rat tag is in new line. fix that.
- [ ] restart server mechanism
  - https://github.com/gin-gonic/gin#graceful-shutdown-or-restart
- CI/CD
  - [x] lint
  - [x] integration tests https://github.com/Orange-OpenSource/hurl#macos
    - [ ] write more
  - [ ] bench mark tests
- store
  - pathutils
    - [ ]
  - cache
    - [ ] delete
    - [ ] load
    - [ ] test
  - fs
  - fs-cache
    - [ ] metadata should be a hidden file
    - [ ] reload at regular interval
- graph
  - [ ] to string - tree view with ids and paths
  - [ ] delete
  - [ ] load
- node
  - rat tags
    - [ ] leafs
    - [ ] img
    - [x] links
      - [ ] can be regular url with name `<rat name url>`
      - [x] link names
      - [x] insert parsed links in template
      - [x] get by id without path
    - [x] graph tree
  - [x] markdown -> html
- http
  - api
    - [ ] auth middleware
    - delete
      - [ ] all - bool flag in body -> DeleteAll
    - [ ] /create/\*path
    - [ ] /read
    - [ ] /update/\*path is an
      - [ ] rename, includes move
      - [ ] cont update
    - [ ] /delete
  - ui
    - [ ] /editor
    - [ ] dark theme
    - [ ] code block seperation
    - [ ] home button
    - editor
      - [x] add node
      - [x] update
      - [x] edit content
      - [x] rename node
      - [ ] clear content
      - [x] delete node
- search
- logging
  - [x] command line flag for log file path
  - [ ] color
  - [ ] setting log level
  - [ ] color and log level settings form config
  - [ ] fix file path of log
- [x] parsing tokens in markdown before converting to html
- [x] func to html
- [x] GraphNode separate file
- [x] command line arguments should be stored in a struct,
  - [x] loader function
- [x] render nodes from route
- [x] wrapper handler func for error
- [x] set up lint
  - [x] fix lint
- [x] config path as argument
- [x] create a handler for reading pages
  - [x] separate router
- [x] use go fumpt, and golines
- [x] config early
- [x] git
- [x] get a server up and running.
- [x] URL path -> node

# links that can be moved

- nodes need to have unique identifier.
  - id field uuid.
- graph stores node as map node id -> node
- graph stores leafs as map node id -> slice of leaf node ids
- graph stores paths as map path -> node id.

# structure

## node

- node is made up of
  - directory - `shell`
  - file in that directory `code`
- properties
  - id uuid
  - name
  - body

## graph

- graph is made up of nodes
- nodes directory can be inside other nodes directory

  - then it is consider a child node

- properties
  - root
  - child node lookup table id -> child nodes

```bash
curl -X POST http://localhost:8888/move \
  -d '{
    "src": "notes/loadero/curls/auth",
    "dest": "notes/loadero/curls/local/auth"
  }'
```
