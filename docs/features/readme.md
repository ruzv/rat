---
id: 29544d51-3b60-44c8-b895-d5c67a96db2d
weight: 1
---

# Readme

<rat graph depth=1 />

---

# Rat

Personal knowledge management tool.

## Description

Bring structure to your personal notes with a feature rich markdown to HTML site
renderer. Take your notes in plain markdown with your editor preference and
allow `rat` to render a pretty version of your notes in you browser.

## Setup

Currently, `rat` has two available options, to get running locally - build from
source or use docker image (recommended as it is the easiest if you already have
a docker setup)

### Docker

Rat docker images are available on
[DockerHub](https://hub.docker.com/r/ruzv/rat). Pull the image with

```sh
docker pull ruzv/rat:latest
```

Then start the container with

```sh
docker run -p 8888:8888 -v /host/dir:/graph ruzv/rat:latest
```

- `-p 8888:8888` - expose the container port 8888 to host port 8888 (Rat server
  in container is configured to serve on port 8888)
- `-v /host/dir:/graph` - mount the host directory `/host/dir` to the container
  directory `/graph` (Rat server in container is configured to use `/graph` as
  the graph directory)

### Build from source

Requirement

- [`go` programming language](https://go.dev/) (version 1.20 or higher is
  recommended, as that's what `rat` is developed with (likely lower version will
  work too, it just has not tested))
- [node.js](https://nodejs.org/en) with npm (again version v20.7.0, because it's
  used for development)

#### Build steps

- Clone this repository
- At the root of the repository there is a `build.sh` script.
  ```sh
  # to produce a binary in current working directory
  ./build.sh -b
  # to install the binary to GOPATH (with go install)
  ./build.sh -i
  ```
- Run
  ```sh
  rat -c config.yaml
  ```

## Configuration

Rat server requires a YAML config file. A minimal config

```yaml
port: 8888
services:
  provider:
    dir: /path/to/graph/dir # directory where to store notes
```

Full config structure available in `config-sample.yaml` file at the root of this
repo.

## Features

- Everything markdown
- File system based, hence
  - Tree like structure of notes
  - Edit notes with in the comfort of your editor
- Automatic sync with git
- Web app
  - Search
  - Create new notes
- Kanban boards
- To-do lists
- Graphviz

## More docs

Currently, Rat has no public documentation site. But to read more about the
specifics of features Rat server offers a documentation site can be started
locally with

```sh
docker run -p 8888:8888 -v $PATH_TO_RAT_REPO/docs:/graph ruzv/rat:latest
```

This command will start a docker container with a rat server hosting all of
Rat's documentation.
