package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

// func (g *Graph) Print() {
// 	callback := func(depth int, gn *graphNode) {
// 		fmt.Printf("%s%s\n", strings.Repeat("   ", depth), gn.node.Name())
// 	}

// 	root := g.Root()

// 	callback(0, root)

// 	err := root.walk(1, callback)
// 	if err != nil {
// 		fmt.Printf("failed to walk graph: %s\n", err)
// 	}
// }

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
		// path:  filepath.Join(g.path, g.nodes[g.root].Name()),
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
		// path:  path,
	}
}

// -------------------------------------------------------------------------- //
// GRAPH NODE
// -------------------------------------------------------------------------- //

type graphNode struct {
	node  *node.Node
	graph *Graph
}

// Leafs returns all leafs of a node. checks cache first. then attempts to load.
// from filesystem. retrieved nodes are cached.
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
		gn, err := n.graph.getByID(id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get leaf")
		}

		gns = append(gns, gn)
	}

	return gns, nil
}

// loads leafs from filesystem and caches them.
func (n *graphNode) loadLeafs() ([]*graphNode, error) {
	fullPath, err := n.fullPath()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get nodes full path")
	}

	logger.Debugf("will load leafs from filesystem %s", fullPath)

	leafs, err := node.Leafs(fullPath)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to load leafs of node %s",
			fullPath,
		)
	}

	path, err := n.path()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get node path")
	}

	gns := make([]*graphNode, 0, len(leafs))
	ids := make([]uuid.UUID, 0, len(leafs))

	for _, leaf := range leafs {
		gns = append(
			gns,
			n.graph.set(leaf, filepath.Join(path, leaf.Name())),
		)

		ids = append(ids, leaf.ID())
	}

	n.graph.leafs[n.node.ID()] = ids

	return gns, nil
}

// TODO: fix path - to full path, and add node as leaf.
func (n *graphNode) Add(name string) (*graphNode, error) {
	path, ok := n.graph.paths[n.node.ID()]
	if !ok {
		return nil, errors.New("node has no path cached")
	}

	path = filepath.Join(path, name)

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

		err = leaf.walk(depth+1, callback)
		if err != nil {
			return errors.Wrap(err, "failed to walk leaf")
		}
	}

	return nil
}

func (n *graphNode) path() (string, error) {
	path, ok := n.graph.paths[n.node.ID()]
	if !ok {
		return "", errors.New("node has no path cached")
	}

	return path, nil
}

func (n *graphNode) fullPath() (string, error) {
	path, err := n.path()
	if err != nil {
		return "", errors.Wrap(err, "failed to get node path")
	}

	return filepath.Join(n.graph.path, path), nil
}

var ratRegex = regexp.MustCompile(
	`<rat( ([[:alnum:]]+)|([[:alnum:]]+)-([[:alnum:]]+))+>`,
)

func ReverseSlice[T any](a []T) []T {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}

	return a
}

func (n *graphNode) Content() string {
	source := n.node.Content()

	matches := ratRegex.FindAllIndex(source, -1)

	parsedSource := string(source)

	for _, match := range ReverseSlice(matches) {
		tag := string(source[match[0]:match[1]])

		parsed, err := n.parseRatTag(tag)
		if err != nil {
			logger.Warnf("failed to parse rat tag %s: %v", tag, err)
		}

		left := parsedSource[:match[0]]
		right := parsedSource[match[1]:]

		parsedSource = left + parsed + right
	}

	return parsedSource
}

// -------------------------------------------------------------------------- //
// RAT TAGS
// -------------------------------------------------------------------------- //

const (
	ratTagKeywordLink = "link"
)

func (n *graphNode) parseRatTag(tag string) (string, error) {
	// <rat keyword arg1 arg2>

	// rat keyword arg1 arg2
	tag = strings.Trim(tag, "<>")

	// keyword arg1 arg2
	args := strings.Split(tag, " ")[1:]

	if len(args) < 1 {
		return "", errors.New("rat tag is missing keyword")
	}

	switch args[0] {
	case ratTagKeywordLink:
		return n.parseRatTagLink(args[1])
	default:
		return "", errors.New("unknown keyword")
	}
}

func (n *graphNode) parseRatTagLink(linkID string) (string, error) {
	id, err := uuid.FromString(linkID)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse link id")
	}

	path, ok := n.graph.paths[id]
	if !ok {
		return "", errors.New("node path not found")
	}

	link := fmt.Sprintf("/%s", path)

	return fmt.Sprintf("<a href=\"%s\">%s</a>", link, path), nil
}
