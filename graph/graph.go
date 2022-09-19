package graph

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"private/rat/errors"
	"private/rat/logger"

	"github.com/gofrs/uuid"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type Store interface {
	GetByID(id uuid.UUID) (*Node, error)
	GetByPath(path string) (*Node, error)
	Leafs(path string) ([]*Node, error)
	Add(parent *Node, name string) (*Node, error)
	Root() (*Node, error)
	Update(node *Node) error
	Move(node *Node, path string) error
	Delete(node *Node) error
}

type Node struct {
	ID      uuid.UUID
	Name    string
	Path    string
	Content string
	Store   Store
}

func (n *Node) Leafs() ([]*Node, error) {
	leafs, err := n.Store.Leafs(n.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get leafs")
	}

	return leafs, nil
}

func (n *Node) GenID() error {
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "failed to generate uuid")
	}

	n.ID = id

	return nil
}

func (n *Node) Add(name string) (*Node, error) {
	node, err := n.Store.Add(n, name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add node")
	}

	return node, nil
}

func renderHook(
	rend *html.Renderer,
) func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	return func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
		b := &bytes.Buffer{}

		var parsed string

		switch n := node.(type) {
		case *ast.CodeBlock:
			rend.CodeBlock(b, n)
			parsed = fmt.Sprintf(
				"<div class=\"markdown-code-block\">%s</div>", b.String(),
			)

		case *ast.Code:
			rend.Code(b, n)
			parsed = fmt.Sprintf(
				"<span class=\"markdown-code\">%s</span>", b.String(),
			)

		default:
			return ast.GoToNext, false
		}

		_, err := io.WriteString(w, parsed)
		if err != nil {
			return ast.GoToNext, false
		}

		return ast.GoToNext, true
	}
}

//nolint:gosec
func (n *Node) HTML() template.HTML {
	return template.HTML(
		markdown.ToHTML(
			[]byte(n.Markdown()),
			parser.NewWithExtensions(parser.CommonExtensions),
			html.NewRenderer(
				html.RendererOptions{
					RenderNodeHook: renderHook(
						html.NewRenderer(html.RendererOptions{}),
					),
				},
			),
		),
	)
}

func (n *Node) Markdown() string {
	matches := ratTagRegex.FindAllIndex([]byte(n.Content), -1)
	parsedSource := n.Content

	for _, match := range ReverseSlice(matches) {
		tag := n.Content[match[0]:match[1]]

		parsed, err := n.parseRatTag(tag)
		if err != nil {
			logger.Warnf("failed to parse rat tag %s: %v", tag, err)
		}

		left := parsedSource[:match[0]]
		right := parsedSource[match[1]:]

		parsedSource = left + parsed + right
	}

	return parsedSource
}

// -------------------------------------------------------------------------- //
// RAT TAGS
// -------------------------------------------------------------------------- //

var ratTagRegex = regexp.MustCompile(
	// `<rat( ([[:alnum:]]+)|([[:alnum:]]+)-([[:alnum:]]+))+>`,
	`<rat(\s([^>]+))>`,
)

func ReverseSlice[T any](a []T) []T {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 { //nolint:lll
		a[left], a[right] = a[right], a[left]
	}

	return a
}

const (
	ratTagKeywordLink  = "link"
	ratTagKeywordGraph = "graph"
	ratTagKeywordImg   = "img"
)

// <rat keyword arg1 arg2> .
func (n *Node) parseRatTag(tag string) (string, error) {
	tag = strings.Trim(tag, "<>") // rat keyword arg1 arg2 .

	// keyword arg1 arg2
	args := strings.Fields(tag)[1:]

	// keyword
	keyword := args[0]

	// arg1 arg2
	args = args[1:]

	switch keyword {
	case ratTagKeywordLink:
		if len(args) == 2 { //nolint:gomnd
			return n.parseRatTagLink(args[0], args[1])
		}

		if len(args) == 1 {
			return n.parseRatTagLink(args[0], "")
		}

		return "", errors.New("too many arguments")
	case ratTagKeywordGraph:
		if len(args) == 0 {
			return n.parseRatTagGraph("-1")
		}

		return n.parseRatTagGraph(args[0])
	case ratTagKeywordImg:
		if len(args) != 1 {
			return "", errors.New("invalid argument count")
		}

		return parseRatTagImg(args[0])
	default:
		return "", errors.New("unknown keyword")
	}
}

func (n *Node) parseRatTagLink(
	linkID string, name string,
) (string, error) {
	id, err := uuid.FromString(linkID)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse link id")
	}

	node, err := n.Store.GetByID(id)
	if err != nil {
		return "", errors.Wrap(err, "failed to get node by ID")
	}

	return node.link(name), nil
}

func (n *Node) parseRatTagGraph(depth string) (string, error) {
	d, err := strconv.Atoi(depth)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse depth")
	}

	// root, err := n.Store.Root()
	// if err != nil {
	// 	return "", errors.Wrap(err, "failed to get root node")
	// }

	var links []string

	err = n.Walk(
		func(depth int, node *Node) bool {
			if depth == d {
				return false
			}

			links = append(
				links,
				fmt.Sprintf(
					"%s- %s", strings.Repeat("\t", depth), node.link(""),
				),
			)

			return true
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to walk graph")
	}

	return strings.Join(links, "\n"), nil
}

func parseRatTagImg(imgPath string) (string, error) {
	imgURL, err := url.JoinPath("/img", imgPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to join img path")
	}

	return fmt.Sprintf("\n<img src=\"%s\">\n", imgURL), nil
}

func (n *Node) link(name string) string {
	path := n.Path

	if name == "" {
		name = n.Name
	}

	return fmt.Sprintf("[%s](/edit/%s)", name, path)
}

func (n *Node) Walk(callback func(int, *Node) bool) error {
	return n.walk(0, callback)
}

func (n *Node) walk(depth int, callback func(int, *Node) bool) error {
	leafs, err := n.Leafs()
	if err != nil {
		return errors.Wrap(err, "failed to get leafs")
	}

	for _, leaf := range leafs {
		if callback != nil && !callback(depth, leaf) {
			continue // callback returned false, skip this branch
		}

		err = leaf.walk(depth+1, callback)
		if err != nil {
			return errors.Wrap(err, "failed to walk leaf")
		}
	}

	return nil
}

// -------------------------------------------------------------------------- //
// UPDATE
// -------------------------------------------------------------------------- //

func (n *Node) Update() error {
	err := n.Store.Update(n)
	if err != nil {
		return errors.Wrap(err, "failed to update")
	}

	return nil
}

func (n *Node) Rename(name string) error {
	err := n.Store.Move(n, filepath.Join(ParentPath(n.Path), name))
	if err != nil {
		return errors.Wrap(err, "failed to rename")
	}

	return nil
}

func (n *Node) Move(path string) error {
	err := n.Store.Move(n, path)
	if err != nil {
		return errors.Wrap(err, "failed to move")
	}

	return nil
}

// -------------------------------------------------------------------------- //
// DELETE
// -------------------------------------------------------------------------- //

func (n *Node) DeleteAll() error {
	err := n.Store.Delete(n)
	if err != nil {
		return errors.Wrap(err, "failed to delete")
	}

	return nil
}

func (n *Node) DeleteSingle() error {
	leafs, err := n.Leafs()
	if err != nil {
		return errors.Wrap(err, "failed to get leafs")
	}

	parent := ParentPath(n.Path)

	for _, leaf := range leafs {
		err = leaf.Move(filepath.Join(parent, leaf.Name))
		if err != nil {
			return errors.Wrap(err, "failed to move leaf node")
		}
	}

	err = n.DeleteAll()
	if err != nil {
		return errors.Wrap(err, "failed to delete all")
	}

	return nil
}

// -------------------------------------------------------------------------- //
// UTILS
// -------------------------------------------------------------------------- //

// NameFromPath returns name of node from its path.
func NameFromPath(path string) string {
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		return ""
	}

	if len(parts) == 1 {
		return parts[0]
	}

	return parts[len(parts)-1]
}

func PathDepth(path string) int {
	return len(strings.Split(path, "/"))
}

func ParentPath(path string) string {
	if PathDepth(path) == 0 {
		return ""
	}

	if PathDepth(path) == 1 {
		return path
	}

	return filepath.Dir(path)
}
