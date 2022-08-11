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

// func (m *metadata) get() error {
// 	metaFilepath := filepath.Join(m.Path, metadataFilename)

// 	f, err := os.Open(metaFilepath)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to open meta file")
// 	}

// 	defer f.Close()

// 	return json.NewDecoder(f).Decode(m)
// }

// func (m *metadata) set() error {
// 	err := os.Mkdir(m.Path, os.ModeDir|os.ModePerm)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to create dir")
// 	}

// 	metaFilepath := filepath.Join(m.Path, metadataFilename)

// 	f, err := os.Create(metaFilepath)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to create meta file")
// 	}

// 	defer f.Close()

// 	return json.NewEncoder(f).Encode(m)
// }

// func (m *metadata) getContent() (*content, error) {
// 	contentPath := filepath.Join(m.Path, contentFilename)

// 	data, err := os.ReadFile(contentPath)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to read content file")
// 	}

// 	return &content{
// 		body: string(data),
// 	}, nil
// }

// func (m *metadata) setContent(c *content) error {
// 	contentPath := filepath.Join(m.Path, contentFilename)

// 	f, err := os.Create(contentPath)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to create content file")
// 	}

// 	defer f.Close()

// 	_, err = f.WriteString(c.body)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to write content")
// 	}

// 	return nil
// }

// func (n *Node) Create() error {
// 	err := n.meta.set()
// 	if err != nil {
// 		return errors.Wrap(err, "failed to set metadata")
// 	}

// 	err = n.meta.setContent(n.cont)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to set content")
// 	}

// 	return nil
// }

// func (n *Node) Read() error {
// 	err := n.meta.get()
// 	if err != nil {
// 		return errors.Wrap(err, "failed to get metadata")
// 	}

// 	n.cont, err = n.meta.getContent()
// 	if err != nil {
// 		return errors.Wrap(err, "failed to get content")
// 	}

// 	return nil
// }

func NameFromPath(path string) string {
	return filepath.Base(path)
}

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

func setMeta(path string, m *metadata) error {
	metaFilepath := filepath.Join(path, metadataFilename)

	f, err := os.Create(metaFilepath)
	if err != nil {
		return errors.Wrap(err, "failed to create meta file")
	}

	defer f.Close()

	return json.NewEncoder(f).Encode(m)
}

func setCont(path string, c *content) error {
	contentPath := filepath.Join(path, contentFilename)

	f, err := os.Create(contentPath)
	if err != nil {
		return errors.Wrap(err, "failed to create content file")
	}

	defer f.Close()

	_, err = f.WriteString(c.body)
	if err != nil {
		return errors.Wrap(err, "failed to write content")
	}

	return nil
}

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

func (n *Node) ID() uuid.UUID {
	return n.meta.ID
}

func (n *Node) Name() string {
	return n.meta.Name
}
