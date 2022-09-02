package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"private/rat/errors"
	"private/rat/graph/node"
	"private/rat/logger"

	"github.com/gofrs/uuid"
)

// Graph holds cached information about graph and its nodes.
type Graph struct {
	path  string                    // path to a directory containing the graph
	root  uuid.UUID                 // uuid of root node
	nodes map[uuid.UUID]*node.Node  // node storage
	leafs map[uuid.UUID][]uuid.UUID // leaf storage
	ids   map[string]uuid.UUID      // path -> node ID
	paths map[uuid.UUID]string      // node ID -> path
}

// Init initializes graph with given name. in given directory (path). if a graph
// with given name already exists, it is loaded.
func Init(name, path string) (*Graph, error) {
	rootPath := filepath.Join(path, name)

	info, err := os.Stat(rootPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "failed to stat root path")
		}

		logger.Infof("will create a new graph with name %s", name)

		return create(rootPath)
	}

	if !info.IsDir() {
		return nil, errors.New("root path is not a directory")
	}

	logger.Infof("will load graph - %s from %s", name, path)

	g := &Graph{
		path:  path,
		nodes: make(map[uuid.UUID]*node.Node),
		leafs: make(map[uuid.UUID][]uuid.UUID),
		ids:   make(map[string]uuid.UUID),
		paths: make(map[uuid.UUID]string),
	}

	n, err := node.Read(rootPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load node")
	}

	g.root = n.ID()
	g.set(n, name)

	err = g.Load()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load graph")
	}

	return g, nil
}

// creates a new graph.
func create(path string) (*Graph, error) {
	g := &Graph{
		path:  path,
		nodes: make(map[uuid.UUID]*node.Node),
		leafs: make(map[uuid.UUID][]uuid.UUID),
		ids:   make(map[string]uuid.UUID),
		paths: make(map[uuid.UUID]string),
	}

	n, err := node.Create(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create root node")
	}

	g.root = n.ID()
	g.set(n, path)

	return g, nil
}

func (g *Graph) Load() error {
	n := g.Root()

	return n.walk(0, nil)
}

func (g *Graph) String() string {
	nodes := make([]string, 0, len(g.nodes))

	for k := range g.nodes {
		var leafs string
		for _, l := range g.leafs[k] {
			leafs = fmt.Sprintf("%s\n\t%s", leafs, l.String())
		}

		nodes = append(
			nodes,
			fmt.Sprintf(
				"id: %s, path: %s\nleafs: %s",
				k.String(),
				g.paths[k],
				leafs,
			),
		)
	}

	return strings.Join(nodes, "\n")
}

// Root returns the root node of the graph.
func (g *Graph) Root() *graphNode { //nolint:golint
	return &graphNode{
		node:  g.nodes[g.root],
		graph: g,
	}
}

// Get returns a node by path. checks cache first. then attempts to load from
// filesystem. retrieved node is cached.
func (g *Graph) Get(path string) (*graphNode, error) { //nolint:golint
	gn, err := g.get(path)
	if err != nil {
		fullPath := filepath.Join(g.path, path)

		n, err := node.Read(fullPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node")
		}

		gn = g.set(n, path)
	}

	return gn, nil
}

// checks cache for node.
func (g *Graph) get(path string) (*graphNode, error) {
	id, ok := g.ids[path]
	if !ok {
		logger.Debugf("node %s not found in cache", path)

		return nil, errors.New("node path not found")
	}

	return g.getByID(id)
}

// check cache for node, error if not found. parentPath path to parent node.
func (g *Graph) getByID(id uuid.UUID) (*graphNode, error) {
	n, ok := g.nodes[id]
	if !ok {
		return nil, errors.New("node not found")
	}

	return &graphNode{
		node:  n,
		graph: g,
	}, nil
}

func (g *Graph) set(n *node.Node, path string) *graphNode {
	g.nodes[n.ID()] = n
	g.ids[path] = n.ID()
	g.paths[n.ID()] = path

	return &graphNode{
		node:  n,
		graph: g,
	}
}
