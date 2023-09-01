package singlefile

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"rat/graph"
	pathutil "rat/graph/util/path"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var (
	markdownFileRe = regexp.MustCompile(`(.*)\.md`)
	headerBodyRe   = regexp.MustCompile(`---\n((?:.|\n)*?)\.\.\.\n((?:.|\n)*)`)
)

var _ graph.Provider = (*SingleFile)(nil)

type nodeHeader struct {
	ID       uuid.UUID `yaml:"id"`
	Template string    `yaml:"template"`
}

// SingleFile is a graph provider implementation that reads and creates graph
// nodes as standalone markdown files as opposed to FileSystem implementation
// creates a content.md .metadata.json and a dir for each node. Maintaining
// the file tree structure of notes.
//
// Example:
// Given a node with a path - notes/tests/integration
// The nodes content would be stored in a file
// rootDir/notes/tests/integration.md
// If the given node has a leaf node "cases", then it's content would be stored
// in - rootDir/notes/tests/integration/cases.md.
type SingleFile struct {
	rootNodePath pathutil.NodePath
	graphPath    string
}

// NewSingleFile creates a new SingleFile graph provider.
func NewSingleFile(
	rootNodePath pathutil.NodePath,
	graphPath string,
) *SingleFile {
	return &SingleFile{
		rootNodePath: rootNodePath,
		graphPath:    graphPath,
	}
}

// GetByID returns a node by its id.
func (sf *SingleFile) GetByID(id uuid.UUID) (*graph.Node, error) {
	r, err := sf.Root()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get root node")
	}

	if r.ID == id {
		return r, nil
	}

	var (
		n     *graph.Node
		found bool
	)

	err = r.Walk(
		sf,
		func(_ int, node *graph.Node) (bool, error) {
			if node.ID == id {
				n = node
				found = true
			}

			return !found, nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk graph")
	}

	if !found {
		return nil, graph.ErrNodeNotFound
	}

	return n, nil
}

// GetByPath returns a node by its path.
func (sf *SingleFile) GetByPath(path pathutil.NodePath) (*graph.Node, error) {
	file, err := os.Open(sf.fullPath(path))
	if err != nil {
		return nil, errors.Wrap(err, "failed to open node markdown file")
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	match := headerBodyRe.FindSubmatch(data)
	if len(match) != 3 {
		return nil, errors.Errorf(
			"failed to match header and body in node %q", string(path),
		)
	}

	header := &nodeHeader{}

	err = yaml.Unmarshal(data, header)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal header as yaml")
	}

	return &graph.Node{
			ID:       header.ID,
			Name:     pathutil.NameFromPath(path),
			Path:     path,
			Content:  string(match[2]),
			Template: header.Template,
		},
		nil
}

// GetLeafs returns all leaf nodes of a node.
func (sf *SingleFile) GetLeafs(path pathutil.NodePath) ([]*graph.Node, error) {
	// graphPath/path/to/node
	leafFiles, err := os.ReadDir(
		filepath.Join(
			sf.graphPath,
			path.String(),
		),
	)
	if err != nil {
		if os.IsNotExist(err) {
			return []*graph.Node{}, nil
		}

		return nil, errors.Wrap(err, "failed to read dir")
	}

	var leafNodes []*graph.Node //nolint:prealloc

	for _, leafFile := range leafFiles {
		if leafFile.IsDir() {
			continue
		}

		match := markdownFileRe.FindStringSubmatch(leafFile.Name())
		if len(match) != 2 {
			continue
		}

		node, err := sf.GetByPath(path.JoinName(match[1]))
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node by path")
		}

		leafNodes = append(leafNodes, node)
	}

	return leafNodes, nil
}

// AddLeaf adds a new leaf to a node.
func (sf *SingleFile) AddLeaf(
	parent *graph.Node, name string,
) (*graph.Node, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate id")
	}

	templ, err := parent.GetTemplate(sf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get template")
	}

	contentBuf := &bytes.Buffer{}

	err = templ.Execute(
		contentBuf,
		struct {
			Name string
		}{
			Name: name,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute template")
	}

	dirPath := filepath.Join(
		sf.graphPath,
		parent.Path.String(),
	)

	err = os.Mkdir(dirPath, 0755)
	if err != nil && !os.IsExist(err) {
		return nil, errors.Wrap(err, "failed to create dir")
	}

	fullPath := filepath.Join(
		sf.graphPath,
		parent.Path.String(),
		fmt.Sprintf("%s.md", name),
	)

	file, err := os.Create(fullPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create file")
	}

	headerData, err := yaml.Marshal(
		nodeHeader{
			ID: id,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal header")
	}

	_, err = file.WriteString(fmt.Sprintf("---\n%s...\n\n", string(headerData)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to write header")
	}

	_, err = file.Write(contentBuf.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "failed to write content")
	}

	return &graph.Node{
			ID:      id,
			Name:    name,
			Path:    parent.Path.JoinName(name),
			Content: contentBuf.String(),
		},
		nil
}

// Move moves a node to a new path.
func (sf *SingleFile) Move(id uuid.UUID, path pathutil.NodePath) error {
	n, err := sf.GetByID(id)
	if err != nil {
		return errors.Wrap(err, "failed to get node by id")
	}

	err = os.Rename(sf.fullPath(n.Path), sf.fullPath(path))
	if err != nil {
		return errors.Wrap(err, "failed to rename node")
	}

	_, err = os.Stat(filepath.Join(sf.graphPath, n.Path.String()))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return errors.Wrap(err, "failed to stat node dir")
	}

	err = os.Rename(
		filepath.Join(sf.graphPath, n.Path.String()),
		filepath.Join(sf.graphPath, path.String()),
	)
	if err != nil {
		return errors.Wrap(err, "failed to rename node dir")
	}

	return nil
}

// Root returns the root node.
func (sf *SingleFile) Root() (*graph.Node, error) {
	return sf.GetByPath(sf.rootNodePath)
}

func (sf *SingleFile) fullPath(path pathutil.NodePath) string {
	if path.Depth() == 1 {
		return filepath.Join(
			sf.graphPath,
			fmt.Sprintf("%s.md", path.NameFromPath()),
		)
	}

	return filepath.Join(
		sf.graphPath,
		path.ParentPath().String(),
		fmt.Sprintf("%s.md", path.NameFromPath()),
	)
}
