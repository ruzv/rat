package graph

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/op/go-logging"
	"github.com/pkg/errors"

	"github.com/gofrs/uuid"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var log = logging.MustGetLogger("graph")

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
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Content string    `json:"content"`
	Store   Store     `json:"-"`
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
	return func(
		w io.Writer, node ast.Node, entering bool,
	) (ast.WalkStatus, bool) {
		// b := &bytes.Buffer{}

		var parsed string

		switch n := node.(type) {
		case *ast.CodeBlock:
			// rend.CodeBlock(b, n)

			lines := strings.Split(string(n.Literal), "\n")

			for idx, line := range lines {
				step := 10

				var (
					parts   []string
					counter int
				)

				for {
					if counter+step > len(line) {
						parts = append(parts, line[counter:])
						break
					}

					parts = append(parts, line[counter:counter+step])
					counter += step
				}

				// for idxP, part := range parts {
				// 	parts[idxP] = fmt.Sprintf("<span>%s</span>", part)
				// }

				lines[idx] = fmt.Sprintf(
					// "<span style=\"display:flex; flex-wrap: wrap;\">%s</span>",
					"<span style=\"flex-wrap: wrap;\">%s</span>",
					strings.Join(parts, "")+"",
				)
			}

			parsed = fmt.Sprintf(
				"<div class=\"markdown-code-block\"><pre><code>%s</code></pre></div>",
				strings.Join(lines, "\n"),
			)

		case *ast.Code:
			parsed = fmt.Sprintf(
				"<span class=\"markdown-code\">%s</span>", string(n.Literal),
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

// -------------------------------------------------------------------------- //
// FORMATS
// -------------------------------------------------------------------------- //

const (
	// nodeFormatRAW      = "raw".
	nodeFormatHTML     = "html"
	nodeFormatMarkdown = "markdown"
)

func (n *Node) HTML() *Node {
	htmlN := *n

	htmlN.Content = string(
		markdown.ToHTML(
			[]byte(n.parseContent(nodeFormatHTML)),
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

	return &htmlN
}

func (n *Node) Markdown() *Node {
	mdN := *n

	mdN.Content = n.parseContent(nodeFormatMarkdown)

	return &mdN
}

// -------------------------------------------------------------------------- //
// RAT TAGS
// -------------------------------------------------------------------------- //

func (n *Node) parseContent(format string) string {
	matches := ratTagRegex.FindAllIndex([]byte(n.Content), -1)
	parsedSource := n.Content

	for _, match := range ReverseSlice(matches) {
		tag := n.Content[match[0]:match[1]]

		parsed, err := n.parseRatTag(tag, format)
		if err != nil {
			log.Warningf("failed to parse rat tag %s: %v", tag, err)
		}

		left := parsedSource[:match[0]]
		right := parsedSource[match[1]:]

		parsedSource = left + parsed + right
	}

	return parsedSource
}

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

//nolint:grouper
const (
	ratTagKeywordLink  = "link"
	ratTagKeywordGraph = "graph"
	ratTagKeywordImg   = "img"
)

// <rat keyword arg1 arg2> .
func (n *Node) parseRatTag(tag, format string) (string, error) {
	pTag := strings.Trim(tag, "<>") // rat keyword arg1 arg2 .

	// keyword arg1 arg2
	args := strings.Fields(pTag)[1:]

	// keyword
	keyword := args[0]

	// arg1 arg2
	args = args[1:]

	switch keyword {
	case ratTagKeywordLink:
		if len(args) == 2 { //nolint:gomnd
			return n.parseRatTagLink(args[0], args[1], format)
		}

		if len(args) == 1 {
			return n.parseRatTagLink(args[0], "", format)
		}

		return "", errors.New("too many arguments")
	case ratTagKeywordGraph:
		if len(args) == 0 {
			return n.parseRatTagGraph("-1", format)
		}

		return n.parseRatTagGraph(args[0], format)
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
	linkID, name, format string,
) (string, error) {
	id, err := uuid.FromString(linkID)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse link id")
	}

	node, err := n.Store.GetByID(id)
	if err != nil {
		return "", errors.Wrap(err, "failed to get node by ID")
	}

	return node.link(name, format), nil
}

func (n *Node) parseRatTagGraph(depth, format string) (string, error) {
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
					"%s- %s",
					strings.Repeat("\t", depth), node.link("", format),
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

func (n *Node) link(name, format string) string {
	var (
		path     = n.Path
		linkName = n.Name
	)

	if name != "" {
		linkName = name
	}

	switch format {
	case nodeFormatHTML:
		var (
			u url.URL
			q = make(url.Values)
		)

		u.Path = "/view/"

		q.Add("node", path)

		u.RawQuery = q.Encode()

		return fmt.Sprintf("[%s](%s)", linkName, u.String())
	case nodeFormatMarkdown:
		return fmt.Sprintf("[%s](/nodes/%s)", linkName, path)
	}

	return "unknown format"
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
