package singlefile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"rat/graph"
	pathutil "rat/graph/util/path"
)

var (
	markdownFileRe = regexp.MustCompile(`(.*)\.md`)
	headerBodyRe   = regexp.MustCompile(`---\n((?:.|\n)*?\n)---\n((?:.|\n)*)`)
)

var _ graph.Provider = (*SingleFile)(nil)

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
) (*SingleFile, error) {
	sf := &SingleFile{
		rootNodePath: rootNodePath,
		graphPath:    graphPath,
	}

	md, _ := sf.fullPath(sf.rootNodePath)

	_, err := os.Stat(md)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "failed to stat root node file")
		}

		err := os.MkdirAll(sf.graphPath, 0o750)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create graph dir")
		}

		id, err := uuid.NewV4()
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate id")
		}

		err = sf.Write(
			&graph.Node{
				Name: string(rootNodePath),
				Path: rootNodePath,
				Header: graph.NodeHeader{
					ID:     id,
					Weight: 0,
					Template: &graph.NodeTemplate{
						Content: "{{ .Name }}",
					},
				},
				Content: "welcome to rat",
			},
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to write root node")
		}
	}

	return sf, nil
}

// GetByID returns a node by its id.
func (sf *SingleFile) GetByID(id uuid.UUID) (*graph.Node, error) {
	r, err := sf.Root()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get root node")
	}

	if r.Header.ID == id {
		return r, nil
	}

	var (
		n     *graph.Node
		found bool
	)

	err = r.Walk(
		sf,
		func(_ int, node *graph.Node) (bool, error) {
			if node.Header.ID == id {
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
		return nil, errors.Wrapf(
			graph.ErrNodeNotFound, "failed to find node by id %s", id.String(),
		)
	}

	return n, nil
}

// GetByPath returns a node by its path.
func (sf *SingleFile) GetByPath(path pathutil.NodePath) (*graph.Node, error) {
	md, _ := sf.fullPath(path)

	file, err := os.Open(md) //nolint:gosec // path cleaned by fullPath
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrapf(
				graph.ErrNodeNotFound,
				"failed to open node markdown file: %s",
				err.Error(),
			)
		}

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

	header := graph.NodeHeader{}

	err = yaml.Unmarshal(data, &header)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal header as yaml")
	}

	return &graph.Node{
			Name:    path.Name(),
			Path:    path,
			Header:  header,
			Content: string(match[2]),
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

	//nolint:prealloc // len(dir entries) != child nodes
	var leafNodes []*graph.Node

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

// Move moves a node to a new path.
func (sf *SingleFile) Move(id uuid.UUID, path pathutil.NodePath) error {
	n, err := sf.GetByID(id)
	if err != nil {
		return errors.Wrap(err, "failed to get node by id")
	}

	oldMD, oldSub := sf.fullPath(n.Path)
	newMD, newSub := sf.fullPath(path)

	err = os.Rename(oldMD, newMD)
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

	err = os.Rename(oldSub, newSub)
	if err != nil {
		return errors.Wrap(err, "failed to rename node dir")
	}

	return nil
}

func (sf *SingleFile) Write(node *graph.Node) error {
	dirPath := sf.graphPath

	if node.Path.Depth() != 1 {
		dirPath = filepath.Join(
			sf.graphPath,
			string(node.Path.ParentPath()),
		)
	}

	err := os.Mkdir(dirPath, 0o750)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "failed to create dir")
	}

	md, _ := sf.fullPath(node.Path)

	file, err := os.Create(md) //nolint:gosec // path cleaned by fullPath
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}

	defer file.Close() //nolint:errcheck // ignore.

	headerData, err := yaml.Marshal(node.Header)
	if err != nil {
		return errors.Wrap(err, "failed to marshal header")
	}

	_, err = fmt.Fprintf(
		file, "---\n%s---\n\n%s", string(headerData), node.Content,
	)
	if err != nil {
		return errors.Wrap(err, "failed to write node file")
	}

	return nil
}

// Delete moves a node and all its sub nodes to the deleted dir.
func (sf *SingleFile) Delete(node *graph.Node) error {
	if node.Path.Depth() == 1 {
		return errors.New("cannot delete root node")
	}

	md, sub := sf.fullPath(node.Path)

	err := os.Remove(md)
	if err != nil {
		return errors.Wrap(err, "failed to remove node markdown file")
	}

	err = os.RemoveAll(sub)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to remove sub nodes dir")
	}

	return nil
}

// Root returns the root node.
func (sf *SingleFile) Root() (*graph.Node, error) {
	return sf.GetByPath(sf.rootNodePath)
}

// returns (markdown, subnode dir) in node tree.
func (sf *SingleFile) fullPath(path pathutil.NodePath) (string, string) {
	if path.Depth() == 1 {
		return filepath.Join(
				sf.graphPath,
				fmt.Sprintf("%s.md", path.Name()),
			),
			filepath.Join(
				sf.graphPath,
				filepath.Clean(path.Name()),
			)
	}

	return filepath.Join(
			sf.graphPath,
			filepath.Clean(path.ParentPath().String()),
			fmt.Sprintf("%s.md", path.Name()),
		),
		filepath.Join(
			sf.graphPath,
			filepath.Clean(path.String()),
		)
}
