package node

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/gofrs/uuid"

	"private/rat/errors"
)

const (
	metadataFilename = "metadata.json"
	contentFilename  = "content.md"
)

// Node is a single part of the graph. Graphs structure is defined by a file
// tree where each node is a directory with files for metadata and content.
// each node can contain subdirectories of other nodes.
type Node struct {
	cont *content
	meta *metadata
}

type metadata struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type content struct {
	body string
}

// NameFromPath returns name of node from its path.
func NameFromPath(path string) string {
	return filepath.Base(path)
}

// Create creates a new Node
func Create(path string) (*Node, error) {
	n := &Node{
		meta: &metadata{
			ID:   uuid.Must(uuid.NewV4()),
			Name: filepath.Base(path),
		},
		cont: &content{},
	}

	err := os.Mkdir(path, os.ModeDir|os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dir")
	}

	err = n.setMeta(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set meta")
	}

	err = n.setCont(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set content")
	}

	return n, nil
}

func (n *Node) setMeta(path string) error {
	metaFilepath := filepath.Join(path, metadataFilename)

	f, err := os.Create(metaFilepath)
	if err != nil {
		return errors.Wrap(err, "failed to create meta file")
	}

	defer f.Close()

	return json.NewEncoder(f).Encode(n.meta)
}

func (n *Node) setCont(path string) error {
	contentPath := filepath.Join(path, contentFilename)

	f, err := os.Create(contentPath)
	if err != nil {
		return errors.Wrap(err, "failed to create content file")
	}

	defer f.Close()

	_, err = f.WriteString(n.cont.body)
	if err != nil {
		return errors.Wrap(err, "failed to write content")
	}

	return nil
}

// Reads a node from filesystem.
func Read(path string) (*Node, error) {
	var (
		node Node
		err  error
	)

	node.meta, err = getMeta(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get meta")
	}

	node.cont, err = getCont(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get content")
	}

	return &node, nil
}

func getMeta(path string) (*metadata, error) {
	metaFilepath := filepath.Join(path, metadataFilename)

	f, err := os.Open(metaFilepath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open meta file")
	}

	defer f.Close()

	var m metadata

	err = json.NewDecoder(f).Decode(&m)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode meta")
	}

	return &m, nil
}

func getCont(path string) (*content, error) {
	contentPath := filepath.Join(path, contentFilename)

	data, err := os.ReadFile(contentPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read content file")
	}

	return &content{
		body: string(data),
	}, nil
}

// Leafs reads all leaf nodes of node specified by path.
func Leafs(path string) ([]*Node, error) {
	leafs, err := os.ReadDir(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read dir %s", path)
	}

	leafNodes := make([]*Node, 0, len(leafs)-2) // cont, meta

	for _, leaf := range leafs {
		if !leaf.IsDir() {
			continue
		}

		path := filepath.Join(path, leaf.Name())

		node, err := Read(path)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node")
		}

		leafNodes = append(leafNodes, node)
	}

	return leafNodes, nil
}

// Node ID.
func (n *Node) ID() uuid.UUID {
	return n.meta.ID
}

// Node Name.
func (n *Node) Name() string {
	return n.meta.Name
}

// Node Content.
func (n *Node) Content() string {
	return n.cont.body
}
