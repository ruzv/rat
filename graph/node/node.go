package node

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type Node struct {
	path string
	name string
	body string
}

type NodeConfig struct {
	Body string `json:"body"`
}

// func (nc *NodeConfig) node() *Node {
// 	return &Node{
// 		body: nc.Body,
// 	}
// }

// func NewNode(name, body string) *Node {
// 	return &Node{
// 		// ID:   uuid.Must(uuid.NewV4()),
// 		// Name: name,
// 		// Body: body,
// 	}
// }

func (n *Node) Path() string {
	return n.path
}

func (n *Node) Name() string {
	return n.name
}

func (n *Node) Body() string {
	return n.body
}

func (n *Node) Add(name string) (*Node, error) {

	// create dir
	// create file
	// write

	path := filepath.Join(n.path, name)

	err := os.Mkdir(path, os.ModeDir|os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dir")
	}

	contentPath := filepath.Join(path, "content.md")

	f, err := os.Create(contentPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create file")
	}

	defer f.Close()

	// err = json.NewEncoder(f).Encode(n.config())
	// if err != nil {
	// 	return "", errors.Wrap(err, "failed to encode node")
	// }

	// n.path = path

	return &Node{
		path: path,
		name: name,
	}, nil
}

// func (n *Node) config() *NodeConfig {
// 	return &NodeConfig{
// 		Body: n.body,
// 	}
// }

// func (n *Node) Save(dir string) (string, error) {
// 	path := filepath.Join(dir, n.Name)

// 	err := os.Mkdir(path, os.ModeDir|os.ModePerm)
// 	if err != nil {
// 		return "", errors.Wrap(err, "failed to create directory")
// 	}

// 	path = filepath.Join(path, "content.md")

// 	f, err := os.Create(path)
// 	if err != nil {
// 		return "", errors.Wrap(err, "failed to create file")
// 	}

// 	defer f.Close()

// 	err = json.NewEncoder(f).Encode(n)
// 	if err != nil {
// 		return "", errors.Wrap(err, "failed to encode node")
// 	}

// 	return path, nil
// }

func Load(path string) (*Node, error) {
	contentPath := filepath.Join(path, "content.md")

	content, err := os.ReadFile(contentPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	var n Node
	n.path = path
	_, n.name = filepath.Split(path)

	if len(content) == 0 {
		return &n, nil
	}

	var nc NodeConfig

	err = json.Unmarshal(content, &nc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal node")
	}

	n.body = nc.Body

	return &n, nil
}

func (n *Node) Leafs() ([]*Node, error) {
	// walk dir
	// find directories
	// load nodes

	leafs, err := os.ReadDir(n.path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read dir")
	}

	leafNodes := make([]*Node, 0, len(leafs)-1)

	for _, leaf := range leafs {
		if !leaf.IsDir() {
			continue
		}

		path := filepath.Join(n.path, leaf.Name())

		n, err := Load(path)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load node")
		}

		leafNodes = append(leafNodes, n)
	}

	return leafNodes, nil
}
