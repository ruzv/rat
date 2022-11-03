package storecache

import (
	"path/filepath"

	"private/rat/graph"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

var _ graph.Store = (*Cache)(nil)

// Cache in memory graph.Store implementation.
type Cache struct {
	root  uuid.UUID
	nodes map[uuid.UUID]*graph.Node // node storage
	ids   map[string]uuid.UUID      // path -> node ID
	paths map[uuid.UUID]string      // node ID -> path
	leafs map[uuid.UUID][]uuid.UUID // leaf storage
}

// GetByID returns node by id.
func (c *Cache) GetByID(id uuid.UUID) (*graph.Node, error) {
	node, ok := c.nodes[id]
	if !ok {
		return nil, errors.New("node not found")
	}

	return node, nil
}

// GetByPath returns node by path.
func (c *Cache) GetByPath(path string) (*graph.Node, error) {
	id, ok := c.ids[path]
	if !ok {
		return nil, errors.New("path not found")
	}

	return c.GetByID(id)
}

// Leafs returns leafs of node.
func (c *Cache) Leafs(path string) ([]*graph.Node, error) {
	node, err := c.GetByPath(path)
	if err != nil {
		return nil, errors.New("path not found")
	}

	leafsIDs, ok := c.leafs[node.ID]
	if !ok {
		return []*graph.Node{}, nil // no leafs
	}

	nodes := make([]*graph.Node, 0, len(leafsIDs))

	for _, id := range leafsIDs {
		n, err := c.GetByID(id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node")
		}

		nodes = append(nodes, n)
	}

	return nodes, nil
}

// Add adds node to cache.
func (c *Cache) Add(parent *graph.Node, name string) (*graph.Node, error) {
	newNode := c.newNode(name, filepath.Join(parent.Path, name))

	err := newNode.GenID()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate id")
	}

	leafs, ok := c.leafs[parent.ID]
	if !ok {
		leafs = []uuid.UUID{}
	}

	leafs = append(leafs, newNode.ID)

	c.ids[newNode.Path] = newNode.ID
	c.paths[newNode.ID] = newNode.Path
	c.leafs[parent.ID] = leafs
	c.nodes[newNode.ID] = newNode

	return newNode, nil
}

// Root returns root node.
func (c *Cache) Root() (*graph.Node, error) {
	return c.GetByID(c.root)
}

func (c *Cache) newNode(name, path string) *graph.Node {
	return &graph.Node{
		Name:  name,
		Path:  path,
		Store: c,
	}
}

// Update updates node.
func (c *Cache) Update(node *graph.Node) error {
	n := c.nodes[node.ID]

	n.Content = node.Content

	c.nodes[node.ID] = n

	return nil
}

// Move moves node to new parent.
func (c *Cache) Move(node *graph.Node, path string) error {
	if c.root == node.ID {
		if graph.PathDepth(path) != 1 {
			return errors.New(
				"cannot move root node, to rename pass path with depth 1",
			)
		}
		// rename root node

		return errors.New("not implemented")
	}

	parent, err := c.parent(node)
	if err != nil {
		return errors.Wrap(err, "failed to get parent")
	}

	c.leafs[parent.ID] = removeFromLeafs(c.leafs[parent.ID], node.ID)
	delete(c.ids, node.Path)
	delete(c.paths, node.ID)

	c.ids[path] = node.ID
	c.paths[node.ID] = path

	node.Path = path
	node.Name = graph.NameFromPath(path)

	parent, err = c.parent(node)
	if err != nil {
		return errors.Wrap(err, "failed to get parent")
	}

	c.leafs[parent.ID] = append(c.leafs[parent.ID], node.ID)

	return nil
}

func (c *Cache) parent(node *graph.Node) (*graph.Node, error) {
	return c.GetByPath(graph.ParentPath(node.Path))
}

func removeFromLeafs(leafs []uuid.UUID, id uuid.UUID) []uuid.UUID {
	removed := make([]uuid.UUID, 0, len(leafs)-1)

	for _, leaf := range leafs {
		if leaf == id {
			continue
		}

		removed = append(removed, leaf)
	}

	return removed
}

// Delete not implemented.
func (*Cache) Delete(_ *graph.Node) error {
	return errors.New("not implemented")
}
