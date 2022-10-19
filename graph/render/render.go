package render

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"private/rat/graph"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/op/go-logging"
	"github.com/pkg/errors"

	"github.com/gofrs/uuid"
)

var log = logging.MustGetLogger("render")

const (
	// nodeFormatRAW      = "raw".
	nodeFormatHTML     = "html"
	nodeFormatMarkdown = "markdown"
)

// -------------------------------------------------------------------------- //
// NODE RENDER
// -------------------------------------------------------------------------- //

type NodeRender struct {
	hookedRend *html.Renderer
	cleanRend  *html.Renderer
}

func NewNodeRender() *NodeRender {
	nr := &NodeRender{
		cleanRend: html.NewRenderer(html.RendererOptions{}),
	}

	nr.hookedRend = html.NewRenderer(html.RendererOptions{
		RenderNodeHook: renderHook(nr),
	})

	return nr
}

func (nr *NodeRender) parser() *parser.Parser {
	return parser.NewWithExtensions(parser.CommonExtensions)
}

// -------------------------------------------------------------------------- //
// HTML
// -------------------------------------------------------------------------- //

func (nr *NodeRender) HTML(n *graph.Node) string {
	return string(markdown.ToHTML(
		[]byte(parseContent(n, nodeFormatHTML)),
		nr.parser(),
		nr.hookedRend,
	))
}

func renderHook(
	rend *NodeRender,
) func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	return func(
		w io.Writer, node ast.Node, entering bool,
	) (ast.WalkStatus, bool) {
		var parsed string

		switch n := node.(type) {
		case *ast.CodeBlock:
			if string(n.Info) == "todo" {
				parsed = renderTodo(rend, n)
			} else {
				parsed = renderCodeBlock(n)
			}

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

func renderCodeBlock(n *ast.CodeBlock) string {
	lines := strings.Split(string(n.Literal), "\n")

	for idx, line := range lines {
		lines[idx] = fmt.Sprintf(
			"<span style=\"flex-wrap: wrap;\">%s</span>",
			line,
		)
	}

	return fmt.Sprintf(
		"<div class=\"markdown-code-block\"><pre><code>%s</code></pre></div>",
		strings.Join(lines, "\n"),
	)
}

func renderTodo(rend *NodeRender, n *ast.CodeBlock) string {
	todoL, err := Parse(string(n.Literal))
	if err != nil {
		// log.Debug("failed to parse", err)

		return ""
	}

	if len(todoL.List) == 0 {
		return ""
	}

	list := append(todoL.NotDone().List, todoL.Done().List...)

	var parts []string

	for _, todo := range list {
		parts = append(parts, todo.render(rend))
	}

	return fmt.Sprintf(
		`<div class="markdown-todo-list">
			%s
		</div>`,
		strings.Join(parts, "\n"),
	)
}

// -------------------------------------------------------------------------- //
// MARKDOWN
// -------------------------------------------------------------------------- //

func (nr *NodeRender) Markdown(n *graph.Node) string {
	return parseContent(n, nodeFormatMarkdown)
}

var ratTagRegex = regexp.MustCompile(
	// `<rat( ([[:alnum:]]+)|([[:alnum:]]+)-([[:alnum:]]+))+>`,
	`<rat(\s([^>]+))>`,
)

func parseContent(n *graph.Node, format string) string {
	matches := ratTagRegex.FindAllIndex([]byte(n.Content), -1)
	parsedSource := n.Content

	for _, match := range ReverseSlice(matches) {
		tag := n.Content[match[0]:match[1]]

		parsed, err := parseRatTag(n, tag, format)
		if err != nil {
			log.Warning("failed to parse rat tag", tag, err)
		}

		left := parsedSource[:match[0]]
		right := parsedSource[match[1]:]

		parsedSource = left + parsed + right
	}

	return parsedSource
}

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
	ratTagKeywordTodo  = "todo"
	ratTagKeywordImg   = "img"
)

// parses a single rat tag
// <rat keyword arg1 arg2> .
func parseRatTag(n *graph.Node, tag, format string) (string, error) {
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
			return parseRatTagLink(n, args[0], args[1], format)
		}

		if len(args) == 1 {
			return parseRatTagLink(n, args[0], "", format)
		}

		return "", errors.New("too many arguments")

	case ratTagKeywordGraph:
		if len(args) == 0 {
			return parseRatTagGraph(n, "-1", format)
		}

		return parseRatTagGraph(n, args[0], format)

	case ratTagKeywordTodo:

		return parseRatTagTodo(n, "", "")

		// case ratTagKeywordImg:
	// 	if len(args) != 1 {
	// 		return "", errors.New("invalid argument count")
	// 	}

	// 	return parseRatTagImg(args[0])
	default:
		return "", errors.New("unknown keyword")
	}
}

// -------------------------------------------------------------------------- //
// link
// -------------------------------------------------------------------------- //

func parseRatTagLink(
	n *graph.Node, linkID, name, format string,
) (string, error) {
	id, err := uuid.FromString(linkID)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse link id")
	}

	node, err := n.Store.GetByID(id)
	if err != nil {
		return "", errors.Wrap(err, "failed to get node by ID")
	}

	return link(node, name, format), nil
}

func link(n *graph.Node, name, format string) string {
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

// -------------------------------------------------------------------------- //
// graph
// -------------------------------------------------------------------------- //

func parseRatTagGraph(n *graph.Node, depth, format string) (string, error) {
	d, err := strconv.Atoi(depth)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse depth")
	}

	var links []string

	err = n.Walk(
		func(depth int, node *graph.Node) bool {
			if depth == d {
				return false
			}

			links = append(
				links,
				fmt.Sprintf(
					"%s- %s",
					strings.Repeat("\t", depth), link(node, "", format),
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

// -------------------------------------------------------------------------- //
// img
// -------------------------------------------------------------------------- //

func parseRatTagImg(imgPath string) (string, error) {
	imgURL, err := url.JoinPath("/img", imgPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to join img path")
	}

	return fmt.Sprintf("\n<img src=\"%s\">\n", imgURL), nil
}

// -------------------------------------------------------------------------- //
// todo
// -------------------------------------------------------------------------- //

var todoRegex = regexp.MustCompile(
	"```todo\n((?:.*\n)*?)```",
)

func parseRatTagTodo(n *graph.Node, depth, format string) (string, error) {
	p, err := n.Parent()
	if err != nil {
		return "", errors.Wrap(err, "failed to get parent")
	}

	var todos []string

	err = p.Walk(
		func(i int, leaf *graph.Node) bool {
			// matches := todoRegex.FindAllString(leaf.Content, -1)
			matches := todoRegex.FindAllStringSubmatch(leaf.Content, -1)
			for _, match := range matches {
				// todos = append(
				// 	todos,
				// 	fmt.Sprintf("### `%s`\n%s", leaf.Path, match),
				// )
				todoL, err := Parse(match[1])
				if err != nil {
					log.Debug("failed to parse todo", err)

					continue
				}

				notDone := todoL.NotDone()
				if len(notDone.List) == 0 {
					continue
				}

				todos = append(
					todos,
					fmt.Sprintf(
						"### `%s`\n%s\n", leaf.Path, notDone.Markdown(),
					),
				)
			}

			return true
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to walk graph")
	}

	// return fmt.Sprintf("```todo\n%s\n```", strings.Join(todos, "")), nil
	return strings.Join(todos, "\n"), nil
}
