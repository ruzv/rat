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
)

// -------------------------------------------------------------------------- //
// GRAPH NODE
// -------------------------------------------------------------------------- //

type graphNode struct {
	node  *node.Node
	graph *Graph
}

// -------------------------------------------------------------------------- //
// LEAFS
// -------------------------------------------------------------------------- //

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

// -------------------------------------------------------------------------- //
// ADD NODE
// -------------------------------------------------------------------------- //

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

var ratTagRegex = regexp.MustCompile(
	`<rat( ([[:alnum:]]+)|([[:alnum:]]+)-([[:alnum:]]+))+>`,
)

func ReverseSlice[T any](a []T) []T {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 { //nolint:lll
		a[left], a[right] = a[right], a[left]
	}

	return a
}

// -------------------------------------------------------------------------- //
// CONTENT
// -------------------------------------------------------------------------- //

func (n *graphNode) Content() string {
	source := n.node.Content()

	matches := ratTagRegex.FindAllIndex(source, -1)

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
