# rat

Personal knowledge management tool.

## Description

Bring structure to your personal notes with a feature rich markdown to HTML site
renderer. Take your notes in plain markdown with your preferenced editor and
allow `rat` to to render a pretty version of your notes in you browser.

## Setup

### Build

Build the go project from source

- Clone this repository
- `cd` into the repository
- Build with `go build` or `go install`
  - `go build` will produce a binary in current working directory
  - `go install` will also produce a binary, but place in your `$GOPATH/bin`

### Run

- Create a directory for your notes
- Configure `rat`
  - A config example file is availible in the root of this repository -
    `config.json`
  - Set the `graph.path` configuration seting to the directory you created for
    your notes
- Run
  - `rat -c config.json`

## Features

### Web App

- Search
  - `cmd + p` to start search
  - `esc` / `i` to enter, exit `insert` mode
  - `n` / `e` to select next / prev search result in `normal` mode (workman
    keyboard layout vim navigation keys, support for `j`, `k` is comming)
  - `enter` to approve result
- Create new node (note)
  - `cmd + shift + p` insert the name of node and approve with `enter`

### Rat Tags

- Todo
- Graph
- Kanban

### Templates

### Markdown -> HTML

### Kanban
