package graph

import (
	"fmt"
	"private/notes/graph/node"
	"strings"

	"github.com/pkg/errors"
)

type Graph struct {
	root  string
	nodes map[string]*node.Node
	leafs map[string][]*node.Node
}

// func NewGraph(dir string, name string) *Graph {
// 	n := node.NewNode(name, "")

// 	_, err := n.Save(dir)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return nil
// }

func Load(root string) (*Graph, error) {
	n, err := node.Load(root)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load node")
	}

	g := &Graph{
		root:  root,
		nodes: map[string]*node.Node{},
		leafs: map[string][]*node.Node{},
	}

	// fmt.Println(n.Path())

	g.nodes[n.Path()] = n

	err = g.Walk()
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk graph")
	}

	return g, nil
}

func (g *Graph) Walk() error {
	n, err := g.Get(g.root)
	if err != nil {
		return errors.Wrap(err, "failed to get root node")
	}

	err = g.walk(n.node, nil, 0)
	if err != nil {
		return errors.Wrap(err, "failed to walk graph")
	}

	return nil
}

func (g *Graph) Print() {
	callback := func(n *node.Node, depth int) {
		fmt.Printf("%s%s\n", strings.Repeat("   ", depth), n.Name())
	}

	n, err := g.Get(g.root)
	if err != nil {
		fmt.Println(err.Error())
	}

	callback(n.Node(), 0)

	err = g.walk(n.Node(), callback, 1)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (g *Graph) walk(n *node.Node, callback func(*node.Node, int), depth int) error {
	children, err := n.Leafs()
	if err != nil {
		return errors.Wrap(err, "failed to get leafs")
	}

	// fmt.Println(n.Name(), len(children), depth)

	g.leafs[n.Path()] = children

	for _, c := range children {
		cc := *c
		child := &cc

		g.nodes[child.Path()] = child

		if callback != nil {
			callback(child, depth)
		}

		err := g.walk(child, callback, depth+1)
		if err != nil {
			return errors.Wrap(err, "failed to walk")
		}
	}

	return nil
}

func (g *Graph) Root() string {
	return g.root
}

type graphNode struct {
	node  *node.Node
	graph *Graph
}

func (g *Graph) Get(path string) (*graphNode, error) {
	n, ok := g.nodes[path]
	if !ok {
		return nil, errors.New("node not found")
	}

	if n == nil {
		return nil, errors.New("found node is nil")
	}

	return &graphNode{
		node:  n,
		graph: g,
	}, nil
}

func (g *Graph) addLeaf(parent string, leaf *node.Node) error {
	// n, err := g.Get(parent)
	// if err != nil {
	// 	return errors.Wrap(err, "failed to get parent")
	// }

	leafs, ok := g.leafs[parent]
	if !ok || leafs == nil {
		leafs = make([]*node.Node, 0, 1)
	}

	leafs = append(leafs, leaf)

	g.leafs[parent] = leafs

	return nil
}

func (gn *graphNode) Node() *node.Node {
	return gn.node
}

func (gn *graphNode) Add(name string) (*graphNode, error) {

	n, err := gn.node.Add(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add node")
	}

	gn.graph.nodes[n.Path()] = n

	err = gn.graph.addLeaf(gn.node.Path(), n)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add leaf")
	}

	return &graphNode{
		node:  n,
		graph: gn.graph,
	}, nil
}
