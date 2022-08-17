package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"private/rat/graph/node"
	"strings"

	"private/rat/errors"

	"github.com/gofrs/uuid"
)

// Graph holds cached information about graph and its nodes.
type Graph struct {
	path  string                    // path to a directory containing the graph
	root  uuid.UUID                 // uuid of root node
	nodes map[uuid.UUID]*node.Node  // node storage
	leafs map[uuid.UUID][]uuid.UUID // leaf storage
	paths map[string]uuid.UUID      // path -> node ID
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

		return create(rootPath)
	}

	if !info.IsDir() {
		return nil, errors.New("root path is not a directory")
	}

	g := &Graph{
		path:  path,
		nodes: make(map[uuid.UUID]*node.Node),
		leafs: make(map[uuid.UUID][]uuid.UUID),
		paths: make(map[string]uuid.UUID),
	}

	n, err := node.Read(rootPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load node")
	}

	g.root = n.ID()
	g.set(n, path)

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
		paths: make(map[string]uuid.UUID),
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

func (g *Graph) Print() {
	callback := func(depth int, gn *graphNode) {
		fmt.Printf("%s%s\n", strings.Repeat("   ", depth), gn.node.Name())
	}

	root := g.Root()

	callback(0, root)

	err := root.walk(1, callback)
	if err != nil {
		fmt.Printf("failed to walk graph: %s\n", err)
	}
}

// returns the root node of the graph
func (g *Graph) Root() *graphNode {
	return &graphNode{
		node:  g.nodes[g.root],
		graph: g,
		path:  filepath.Join(g.path, g.nodes[g.root].Name()),
	}
}

// Get returns a node by path. checks cache first. then attempts to load from
// filesystem. retrived node is cached.
func (g *Graph) Get(path string) (*graphNode, error) {
	path = filepath.Join(g.path, path)

	gn, err := g.get(path)
	if err != nil {
		n, err := node.Read(path)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node")
		}

		gn = g.set(n, path)
	}

	return gn, nil
}

// checks cache for node
func (g *Graph) get(path string) (*graphNode, error) {
	id, ok := g.paths[path]
	if !ok {
		return nil, errors.New("node path not found")
	}

	return g.getByID(id, filepath.Dir(path))
}

// check cache for node, error if not found
func (g *Graph) getByID(id uuid.UUID, parentPath string) (*graphNode, error) {
	n, ok := g.nodes[id]
	if !ok {
		return nil, errors.New("node not found")
	}

	return &graphNode{
		node:  n,
		graph: g,
		path:  filepath.Join(parentPath, n.Name()),
	}, nil
}

func (g *Graph) set(n *node.Node, path string) *graphNode {
	g.nodes[n.ID()] = n
	g.paths[path] = n.ID()

	return &graphNode{
		node:  n,
		graph: g,
		path:  path,
	}
}

// -------------------------------------------------------------------------- //
// GRAPH NODE
// -------------------------------------------------------------------------- //

type graphNode struct {
	node  *node.Node
	graph *Graph
	path  string
}

// Leafs returns all leafs of a node. checks cache first. then attempts to load.
// from filesystem. retrived nodes are cached.
func (n *graphNode) Leafs() ([]*graphNode, error) {
	leafs, err := n.cachedLeafs()
	if err != nil {

		leafs, err = n.loadLeafs()
		if err != nil {
			return nil, errors.Wrap(err, "failed to load leafs")
		}
	}

	return leafs, nil
}

// leafs checks graphs cache, errors if not found.
func (n *graphNode) cachedLeafs() ([]*graphNode, error) {
	leafsIDs, ok := n.graph.leafs[n.node.ID()]
	if !ok {
		return nil, errors.New("node has no leafs cached")
	}

	gns := make([]*graphNode, 0, len(leafsIDs))

	for _, id := range leafsIDs {
		gn, err := n.graph.getByID(id, n.node.Name())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get leaf")
		}

		gns = append(gns, gn)
	}

	return gns, nil
}

// loads leafs from filesystem and caches them.
func (n *graphNode) loadLeafs() ([]*graphNode, error) {
	leafs, err := node.Leafs(n.path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load leafs of node %s", n.path)
	}

	gns := make([]*graphNode, 0, len(leafs))
	for _, leaf := range leafs {
		gns = append(
			gns,
			n.graph.set(leaf, filepath.Join(n.path, leaf.Name())),
		)
	}

	return gns, nil
}

func (n *graphNode) Add(name string) (*graphNode, error) {
	path := filepath.Join(n.path, name)

	newNode, err := node.Create(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create node")
	}

	gn := n.graph.set(newNode, path)

	return gn, nil
}

// traverses the graph starting from node n. caching newly loaded nodes.
func (n *graphNode) walk(
	depth int, callback func(int, *graphNode),
) error {
	leafs, err := n.Leafs()
	if err != nil {
		return errors.Wrap(err, "failed to get leafs")
	}

	for _, leaf := range leafs {
		if callback != nil {
			callback(depth, leaf)
		}

		leaf.walk(depth+1, callback)
	}

	return nil
}

func (n *graphNode) Content() string {
	return n.node.Content()
}
