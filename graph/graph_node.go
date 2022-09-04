package graph

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"private/rat/errors"
	"private/rat/graph/node"
	"private/rat/logger"

	"github.com/gofrs/uuid"
	"github.com/gomarkdown/markdown"
)

// -------------------------------------------------------------------------- //
// GRAPH NODE
// -------------------------------------------------------------------------- //

type GraphNode struct {
	node  *node.Node
	graph *Graph
}

// -------------------------------------------------------------------------- //
// LEAFS
// -------------------------------------------------------------------------- //

// Leafs returns all leafs of a node. checks cache first. then attempts to load.
// from filesystem. retrieved nodes are cached.
func (n *GraphNode) Leafs() ([]*GraphNode, error) {
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
func (n *GraphNode) cachedLeafs() ([]*GraphNode, error) {
	leafsIDs, ok := n.graph.leafs[n.node.ID()]
	if !ok {
		return nil, errors.New("node has no leafs cached")
	}

	gns := make([]*GraphNode, 0, len(leafsIDs))

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
func (n *GraphNode) loadLeafs() ([]*GraphNode, error) {
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

	gns := make([]*GraphNode, 0, len(leafs))
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

// -------------------------------------------------------------------------- //
// ADD NODE
// -------------------------------------------------------------------------- //

// TODO: fix path - to full path, and add node as leaf.
func (n *GraphNode) Add(name string) (*GraphNode, error) {
	parentFullPath, err := n.fullPath()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get full path")
	}

	parentPath, err := n.path()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get path")
	}

	newNode, err := node.Create(filepath.Join(parentFullPath, name))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create node")
	}

	gn := n.graph.set(newNode, filepath.Join(parentPath, name))

	n.addLeaf(gn)

	return gn, nil
}

func (n *GraphNode) addLeaf(node *GraphNode) {
	leafs, ok := n.graph.leafs[n.node.ID()]
	if !ok {
		leafs = []uuid.UUID{}
	}

	leafs = append(leafs, node.node.ID())

	n.graph.leafs[n.node.ID()] = leafs
}

// traverses the graph starting from node n. caching newly loaded nodes.
func (n *GraphNode) walk(
	depth int, callback func(int, *GraphNode),
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

func (n *GraphNode) path() (string, error) {
	path, ok := n.graph.paths[n.node.ID()]
	if !ok {
		return "", errors.New("node has no path cached")
	}

	return path, nil
}

func (n *GraphNode) fullPath() (string, error) {
	path, err := n.path()
	if err != nil {
		return "", errors.Wrap(err, "failed to get node path")
	}

	return filepath.Join(n.graph.path, path), nil
}

// Path retrives the nodes path.
func (n *GraphNode) Path() string {
	path, err := n.path()
	if err != nil {
		return ""
	}

	return path
}

func (n *GraphNode) Name() string {
	return node.NameFromPath(n.Path())
}

// -------------------------------------------------------------------------- //
// CONTENT
// -------------------------------------------------------------------------- //

type GraphNodeContent struct {
	graphNode *GraphNode
}

func (n *GraphNode) Content() *GraphNodeContent {
	return &GraphNodeContent{
		graphNode: n,
	}
}

func (c *GraphNodeContent) Markdown() string {
	source := c.graphNode.node.Content()

	matches := ratTagRegex.FindAllIndex(source, -1)

	parsedSource := string(source)

	for _, match := range ReverseSlice(matches) {
		tag := string(source[match[0]:match[1]])

		parsed, err := c.graphNode.parseRatTag(tag)
		if err != nil {
			logger.Warnf("failed to parse rat tag %s: %v", tag, err)
		}

		left := parsedSource[:match[0]]
		right := parsedSource[match[1]:]

		parsedSource = left + parsed + right
	}

	return parsedSource
}

func (c *GraphNodeContent) HTML() string {
	return string(markdown.ToHTML([]byte(c.Markdown()), nil, nil))
}

func (c *GraphNodeContent) Raw() string {
	return string(c.graphNode.node.Content())
}

// -------------------------------------------------------------------------- //
// UPDATE
// -------------------------------------------------------------------------- //

type GraphNodeUpdate struct {
	graphNode *GraphNode
}

func (n *GraphNode) Update() *GraphNodeUpdate {
	return &GraphNodeUpdate{
		graphNode: n,
	}
}

func (u *GraphNodeUpdate) Name(name string) error {
	path, err := u.graphNode.path()
	if err != nil {
		return errors.Wrap(err, "failed to get node path")
	}

	fullPath, err := u.graphNode.fullPath()
	if err != nil {
		return errors.Wrap(err, "failed to get node full path")
	}

	newPath, err := u.graphNode.node.Rename(fullPath, name)
	if err != nil {
		return errors.Wrap(err, "failed to rename node in filesystem")
	}

	delete(u.graphNode.graph.ids, path)

	u.graphNode.graph.ids[newPath] = u.graphNode.node.ID()

	u.graphNode.graph.paths[u.graphNode.node.ID()] = newPath

	return nil
}

func (u *GraphNodeUpdate) Content(content string) error {
	fullPath, err := u.graphNode.fullPath()
	if err != nil {
		return errors.Wrap(err, "failed to get full path")
	}

	err = u.graphNode.node.Update(fullPath, content)
	if err != nil {
		return errors.Wrap(err, "failed to update node content")
	}

	return nil
}

// -------------------------------------------------------------------------- //
// RAT TAGS
// -------------------------------------------------------------------------- //

var ratTagRegex = regexp.MustCompile(
	`<rat( ([[:alnum:]]+)|([[:alnum:]]+)-([[:alnum:]]+))+>`,
)

func ReverseSlice[T any](a []T) []T {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 { //nolint:lll
		a[left], a[right] = a[right], a[left]
	}

	return a
}

const (
	ratTagKeywordLink  = "link"
	ratTagKeywordGraph = "graph"
)

func (n *GraphNode) parseRatTag(tag string) (string, error) {
	// <rat keyword arg1 arg2>

	// rat keyword arg1 arg2
	tag = strings.Trim(tag, "<>")

	// keyword arg1 arg2
	args := strings.Split(tag, " ")[1:]

	// keyword
	keyword := args[0]

	// arg1 arg2
	args = args[1:]

	switch keyword {
	case ratTagKeywordLink:
		if len(args) == 2 {
			return n.parseRatTagLink(args[0], args[1])
		}

		if len(args) == 1 {
			return n.parseRatTagLink(args[0], "")
		}

		return "", errors.New("too many arguments")
	case ratTagKeywordGraph:
		return n.parseRatTagGraph()
	default:
		return "", errors.New("unknown keyword")
	}
}

func (n *GraphNode) parseRatTagLink(
	linkID string, name string,
) (string, error) {
	id, err := uuid.FromString(linkID)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse link id")
	}

	node, err := n.graph.getByID(id)
	if err != nil {
		return "", errors.Wrap(err, "failed to get node by ID")
	}

	return node.link(name), nil
}

func (n *GraphNode) parseRatTagGraph() (string, error) {
	var links []string

	err := n.graph.Root().walk(
		0,
		func(depth int, gn *GraphNode) {
			links = append(
				links,
				fmt.Sprintf("%s- %s", strings.Repeat("\t", depth), gn.link("")),
			)
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to walk graph")
	}

	return strings.Join(links, "\n"), nil
}

func (n *GraphNode) link(name string) string {
	path := n.Path()

	if name == "" {
		name = n.Name()
	}

	return fmt.Sprintf("[%s](/graphs/%s)", name, path)
}
