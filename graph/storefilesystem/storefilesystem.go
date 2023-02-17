package storefilesystem

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"private/rat/graph"
	pathutil "private/rat/graph/util/path"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

var _ graph.Provider = (*FileSystem)(nil)

const (
	metadataFilename = ".metadata.json"
	contentFilename  = "content.md"
	templateFilename = ".template.md"
)

// FileSystem implements graph.Store reading and writing to the filesystem. for
// every operation.
type FileSystem struct {
	root string // node path, not full path
	path string // path to a directory containing the graph
}

// NewFileSystem returns a new FileSystem.
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

// GetByID returns a node by ID.
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
		fs,
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

// GetByPath returns a node by path.
func (fs *FileSystem) GetByPath(path string) (*graph.Node, error) {
	fullpath := fs.fullPath(path)

	node := fs.newNode(pathutil.NameFromPath(path), path)

	err := getMeta(node, fullpath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get meta")
	}

	err = getCont(node, fullpath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cont")
	}

	err = getTemplate(node, fullpath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get template")
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

func getTemplate(node *graph.Node, path string) error {
	p := filepath.Join(path, templateFilename)

	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}

	node.Template = string(data)

	return nil
}

func (fs *FileSystem) newNode(name, path string) *graph.Node {
	return &graph.Node{
		Name: name,
		Path: path,
	}
}

func (fs *FileSystem) fullPath(path string) string {
	return filepath.Join(fs.path, path)
}

// GetLeafs returns leaf nodes.
func (fs *FileSystem) GetLeafs(id uuid.UUID) ([]*graph.Node, error) {
	parent, err := fs.GetByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get parent node")
	}

	fullPath := fs.fullPath(parent.Path)

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

		node, err := fs.GetByPath(filepath.Join(parent.Path, leaf.Name()))
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node")
		}

		leafNodes = append(leafNodes, node)
	}

	return leafNodes, nil
}

// Add adds a new node to the graph.
func (fs *FileSystem) AddLeaf(
	parent *graph.Node, name string,
) (*graph.Node, error) {
	newNode := fs.newNode(name, filepath.Join(parent.Path, name))
	newFullPath := filepath.Join(fs.fullPath(parent.Path), name)

	templ, err := parent.GetTemplate(fs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get template")
	}

	buff := &bytes.Buffer{}

	err = templ.Execute(
		buff,
		struct {
			Name string
		}{
			Name: name,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute template")
	}

	newNode.Content = buff.String()

	newNode.ID, err = uuid.NewV4()
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

// Root returns the root node of the graph.
func (fs *FileSystem) Root() (*graph.Node, error) {
	return fs.GetByPath(fs.root)
}
