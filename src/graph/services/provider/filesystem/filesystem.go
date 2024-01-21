package filesystem

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
	"rat/logr"
)

var (
	markdownFileRe = regexp.MustCompile(`(.*)\.md`)
	headerBodyRe   = regexp.MustCompile(`---\n((?:.|\n)*?\n)---\n((?:.|\n)*)`)
)

var _ graph.RootsProvider = (*Provider)(nil)

// Provider is a graph provider implementation that reads and creates graph
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
type Provider struct {
	graphDir string
}

// NewProvider creates a new filesystem graph provider.
func NewProvider(graphDir string, log *logr.LogR) (*Provider, error) {
	log = log.Prefix("filesystem")

	_, err := os.Stat(graphDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(
				err, "failed to get status of graph directory",
			)
		}

		err := os.MkdirAll(graphDir, 0o750)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create graph dir")
		}

		log.Infof("created %q dir", graphDir)
	}

	log.Infof("dir %q", graphDir)

	return &Provider{
		graphDir: graphDir,
	}, nil
}

// Roots returns all nodes with depth 1.
func (p *Provider) Roots() ([]*graph.Node, error) {
	graphDirEntries, err := os.ReadDir(p.graphDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read graph dir")
	}

	//nolint:prealloc // len(dir entries) != child nodes
	var firstNodes []*graph.Node

	for _, graphDirEntry := range graphDirEntries {
		if graphDirEntry.IsDir() {
			continue
		}

		match := markdownFileRe.FindStringSubmatch(graphDirEntry.Name())
		if len(match) != 2 {
			continue
		}

		node, err := p.GetByPath(pathutil.NodePath(match[1]))
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node by path")
		}

		firstNodes = append(firstNodes, node)
	}

	return firstNodes, nil
}

// GetByID returns a node by its id.
func (p *Provider) GetByID(id uuid.UUID) (*graph.Node, error) {
	roots, err := p.GetLeafs("")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get first nodes")
	}

	for _, root := range roots {
		if root.Header.ID == id {
			return root, nil
		}

		var (
			n     *graph.Node
			found bool
		)

		err := root.Walk(
			p,
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
			continue
		}

		return n, nil
	}

	return nil, errors.Wrapf(
		graph.ErrNodeNotFound, "failed to find node by id %s", id.String(),
	)
}

// GetByPath returns a node by its path. If the markdown file does not have a
// header or header is missing an ID, it is added to the file.
//
//nolint:gocyclo,cyclop // 11, not that big
func (p *Provider) GetByPath(path pathutil.NodePath) (*graph.Node, error) {
	md, _ := p.fullPath(path)

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

	n := &graph.Node{
		Path: path,
	}

	match := headerBodyRe.FindSubmatch(data)
	if len(match) != 3 {
		n.Content = string(data)

		_, err := n.FillID()
		if err != nil {
			return nil, errors.Wrapf(
				err, "failed to fill id, for node %q without header", path,
			)
		}

		err = p.Write(n)
		if err != nil {
			return nil, errors.Wrapf(
				err, "failed create header for node %q", path,
			)
		}

		return n, nil
	}

	n.Content = string(match[2])

	header := graph.NodeHeader{}

	err = yaml.Unmarshal(data, &header)
	if err != nil {
		return nil, errors.Wrapf(
			err, "failed to unmarshal header as yaml for node %q", path,
		)
	}

	n.Header = header

	if n.Header.ID.IsNil() {
		_, err := n.FillID()
		if err != nil {
			return nil, errors.Wrapf(
				err, "failed to fill id, for node %q without header", path,
			)
		}

		err = p.Write(n)
		if err != nil {
			return nil, errors.Wrapf(
				err, "failed create header for node %q", path,
			)
		}

		return n, nil
	}

	return n, nil
}

// GetLeafs returns all leaf nodes of a node.
func (p *Provider) GetLeafs(path pathutil.NodePath) ([]*graph.Node, error) {
	_, subDir := p.fullPath(path)

	// graphPath/path/to/node
	leafFiles, err := os.ReadDir(subDir)
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

		node, err := p.GetByPath(path.JoinName(match[1]))
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node by path")
		}

		leafNodes = append(leafNodes, node)
	}

	return leafNodes, nil
}

// Move moves a node to a new path.
func (p *Provider) Move(id uuid.UUID, path pathutil.NodePath) error {
	// check if dest parent exists
	_, err := p.GetByPath(path.Parent())
	if err != nil {
		return errors.Wrap(
			err,
			"failed to get parent node of new path, parent must already exist",
		)
	}

	// check if move target exists
	n, err := p.GetByID(id)
	if err != nil {
		return errors.Wrap(err, "failed to get node by id")
	}

	oldMD, oldSub := p.fullPath(n.Path)
	newMD, newSub := p.fullPath(path)

	// move a/b/c
	//   to a/k/r
	// a/k exists, but no a/k dir for sub nodes
	// a/k.md

	// check if dest parent has sub dir
	destParentSubDir := filepath.Dir(newSub)

	_, err = os.Stat(destParentSubDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to stat new node dir")
		}

		err = os.MkdirAll(destParentSubDir, 0o750)
		if err != nil {
			return errors.Wrap(err, "failed to create new node parent sub dir")
		}
	}

	err = os.Rename(oldMD, newMD)
	if err != nil {
		return errors.Wrap(err, "failed to rename node")
	}

	_, err = os.Stat(oldSub)
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

// Write writes a node to a markdown file. If file already exists it is
// overwritten.
func (p *Provider) Write(node *graph.Node) error {
	md, _ := p.fullPath(node.Path)

	err := os.Mkdir(filepath.Dir(md), 0o750)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "failed to create dir")
	}

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
func (p *Provider) Delete(node *graph.Node) error {
	md, sub := p.fullPath(node.Path)

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

// returns (markdown, subnode dir) in node tree.
func (p *Provider) fullPath(path pathutil.NodePath) (string, string) {
	if path.Depth() == 1 {
		return filepath.Join(
				p.graphDir,
				fmt.Sprintf("%s.md", path.Name()),
			),
			filepath.Join(
				p.graphDir,
				filepath.Clean(path.Name()),
			)
	}

	return filepath.Join(
			p.graphDir,
			filepath.Clean(path.Parent().String()),
			fmt.Sprintf("%s.md", path.Name()),
		),
		filepath.Join(
			p.graphDir,
			filepath.Clean(path.String()),
		)
}
