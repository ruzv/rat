package storefilesystem

import (
	"encoding/json"
	"os"
	"path/filepath"

	"private/rat/errors"
	"private/rat/graph"

	"github.com/gofrs/uuid"
)

var _ graph.Store = (*FileSystem)(nil)

const (
	metadataFilename = ".metadata.json"
	contentFilename  = "content.md"
)

type FileSystem struct {
	root string // node path, not full path
	path string // path to a directory containing the graph
}

func NewFileSystem(root, path string) (*FileSystem, error) {
	fs := &FileSystem{
		root: root,
		path: path,
	}

	dir, err := os.ReadDir(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read dir")
	}

	for _, d := range dir {
		if d.Name() == root {
			if !d.IsDir() {
				return nil, errors.New("root node named file")
			}

			return fs, nil
		}
	}

	p := filepath.Join(path, root)

	err = os.MkdirAll(p, os.ModePerm)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create root node dir")
	}

	n := fs.newNode(root, root)

	err = setCont(n, p)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to set cont")
	}

	err = setMeta(n, p)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to set meta")
	}

	return fs, nil
}

func (fs *FileSystem) GetByID(id uuid.UUID) (*graph.Node, error) {
	root, err := fs.Root()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get root node")
	}

	if root.ID == id {
		return root, nil
	}

	var (
		n     *graph.Node
		found bool
	)

	err = root.Walk(
		func(_ int, node *graph.Node) bool {
			if node.ID == id {
				n = node
				found = true
			}

			return !found
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk graph")
	}

	return n, nil
}

func (fs *FileSystem) GetByPath(path string) (*graph.Node, error) {
	fullpath := fs.fullPath(path)

	node := fs.newNode(graph.NameFromPath(path), path)

	err := getMeta(node, fullpath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get meta")
	}

	err = getCont(node, fullpath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cont")
	}

	return node, nil
}

func getMeta(node *graph.Node, path string) error {
	metaFilepath := filepath.Join(path, metadataFilename)

	f, err := os.Open(metaFilepath)
	if err != nil {
		return errors.Wrap(err, "failed to open meta file")
	}

	defer f.Close()

	var m struct {
		ID uuid.UUID `json:"id"`
	}

	err = json.NewDecoder(f).Decode(&m)
	if err != nil {
		return errors.Wrap(err, "failed to decode meta")
	}

	node.ID = m.ID

	// m.name = NameFromPath(path)

	return nil
}

func getCont(node *graph.Node, path string) error {
	contentPath := filepath.Join(path, contentFilename)

	data, err := os.ReadFile(contentPath)
	if err != nil {
		return errors.Wrap(err, "failed to read content file")
	}

	node.Content = string(data)

	return nil
}

func (fs *FileSystem) newNode(name, path string) *graph.Node {
	return &graph.Node{
		Name:  name,
		Path:  path,
		Store: fs,
	}
}

func (fs *FileSystem) fullPath(path string) string {
	return filepath.Join(fs.path, path)
}

func (fs *FileSystem) Leafs(path string) ([]*graph.Node, error) {
	fullPath := fs.fullPath(path)

	leafs, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read dir %s", fullPath)
	}

	// cont, meta
	leafNodes := make([]*graph.Node, 0, len(leafs)-2) //nolint:gomnd

	for _, leaf := range leafs {
		if !leaf.IsDir() {
			continue
		}

		node, err := fs.GetByPath(filepath.Join(path, leaf.Name()))
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node")
		}

		leafNodes = append(leafNodes, node)
	}

	return leafNodes, nil
}

func (fs *FileSystem) Add(
	parent *graph.Node,
	name string,
) (*graph.Node, error) {
	newNode := fs.newNode(name, filepath.Join(parent.Path, name))
	newFullPath := filepath.Join(fs.fullPath(parent.Path), name)

	err := newNode.GenID()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate id")
	}

	err = os.Mkdir(newFullPath, os.ModeDir|os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dir")
	}

	err = setMeta(newNode, newFullPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set meta")
	}

	err = setCont(newNode, newFullPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set content")
	}

	return newNode, nil
}

func setMeta(node *graph.Node, path string) error {
	metaFilepath := filepath.Join(path, metadataFilename)

	f, err := os.Create(metaFilepath)
	if err != nil {
		return errors.Wrap(err, "failed to create meta file")
	}

	defer f.Close()

	e := json.NewEncoder(f)

	e.SetIndent("", "    ")

	m := struct {
		ID uuid.UUID `json:"id"`
	}{
		ID: node.ID,
	}

	err = e.Encode(m)
	if err != nil {
		return errors.Wrap(err, "failed to encode meta")
	}

	return nil
}

func setCont(node *graph.Node, path string) error {
	contentPath := filepath.Join(path, contentFilename)

	f, err := os.Create(contentPath)
	if err != nil {
		return errors.Wrap(err, "failed to create content file")
	}

	defer f.Close()

	_, err = f.Write([]byte(node.Content))
	if err != nil {
		return errors.Wrap(err, "failed to write content")
	}

	return nil
}

func (fs *FileSystem) Root() (*graph.Node, error) {
	return fs.GetByPath(fs.root)
}

func (fs *FileSystem) Update(node *graph.Node) error {
	return setCont(node, fs.fullPath(node.Path))
}

func (fs *FileSystem) Move(node *graph.Node, path string) error {
	err := os.Rename(fs.fullPath(node.Path), fs.fullPath(path))
	if err != nil {
		return errors.Wrap(err, "failed to move node")
	}

	node.Path = path
	node.Name = graph.NameFromPath(path)

	return nil
}

func (fs *FileSystem) Delete(node *graph.Node) error {
	err := os.RemoveAll(fs.fullPath(node.Path))
	if err != nil {
		return errors.Wrapf(err, "failed to remove all")
	}

	return nil
}
