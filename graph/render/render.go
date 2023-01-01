package render

import (
	"fmt"
	"io"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"private/rat/graph"
	"private/rat/graph/render/todo"

	"github.com/gofrs/uuid"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var log = logging.MustGetLogger("render")

type linkFormat string

const (
	linkFormatWeb = "web"
	linkFormatAPI = "api"
)

// -------------------------------------------------------------------------- //
// NODE RENDER
// -------------------------------------------------------------------------- //

// NodeRender groups renderers used to render nodes from markdown to html.
type NodeRender struct {
	hookedRend *html.Renderer
	cleanRend  *html.Renderer
}

// NewNodeRender creates a new NodeRender.
func NewNodeRender() *NodeRender {
	nr := &NodeRender{
		cleanRend: html.NewRenderer(html.RendererOptions{}),
	}

	nr.hookedRend = html.NewRenderer(html.RendererOptions{
		RenderNodeHook: renderHook(nr),
	})

	return nr
}

func (*NodeRender) parser() *parser.Parser {
	return parser.NewWithExtensions(parser.CommonExtensions)
}

// -------------------------------------------------------------------------- //
// HTML
// -------------------------------------------------------------------------- //

// HTML renders the node as HTML string.
func (nr *NodeRender) HTML(n *graph.Node) string {
	return string(markdown.ToHTML(
		[]byte(parseContent(n, linkFormatWeb)),
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
				parsed = renderTodoList(rend, n)
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

func renderTodoList(rend *NodeRender, n *ast.CodeBlock) string {
	todoL, err := todo.Parse(string(n.Literal))
	if err != nil {
		log.Warning("failed to parse todo list", err)

		return renderError(errors.Wrap(err, "failed to parse todo"))
	}

	if todoL.Empty() {
		return ""
	}

	list := append(todoL.NotDone().List, todoL.Done().List...) //nolint:gocritic
	parts := make([]string, 0, len(list))

	for _, todo := range list {
		parts = append(parts, renderTodo(rend, todo))
	}

	var header []string
	header = append(header, `<div class="markdown-todo-params">`)

	if len(todoL.SourceNodePath) != 0 {
		header = append(
			header,
			fmt.Sprintf(
				`<span class="markdown-code"><a href="%s">%s</a></span>`,
				link(todoL.SourceNodePath, linkFormatWeb),
				todoL.SourceNodePath,
			),
		)
	}

	if len(todoL.PriorityString()) != 0 {
		header = append(
			header,
			fmt.Sprintf(
				`<div><span class="markdown-code">%s</span></div>`,
				todoL.PriorityString(),
			),
		)
	}

	if len(todoL.DueString()) != 0 {
		header = append(
			header,
			fmt.Sprintf(
				`<div><span class="markdown-code">%s</span></div>`,
				todoL.DueString(),
			),
		)
	}

	if len(todoL.SizeString()) != 0 {
		header = append(
			header,
			fmt.Sprintf(
				`<div><span class="markdown-code">%s</span></div>`,
				todoL.SizeString(),
			),
		)
	}

	header = append(header, `</div>`)

	return fmt.Sprintf(
		`<div class="markdown-todo-list">
			<div class="markdown-todo-header">
				%s
			</div>
			%s
		</div>`,
		strings.Join(header, "\n"),
		strings.Join(parts, "\n"),
	)
}

func renderTodo(rend *NodeRender, td todo.Todo) string {
	var checked string

	if td.Done {
		checked = " markdown-todo-checkbox-check-checked"
	}

	return fmt.Sprintf(
		`<div class="markdown-todo">
			<div class="markdown-todo-checkbox-border" onclick="todoDone(event)">
				<div class="markdown-todo-checkbox-check%s">
				</div>
			</div>			
			<div class="markdown-todo-text">
				%s
			</div>
		</div>`,
		checked,
		markdown.ToHTML(
			[]byte(td.Text),
			rend.parser(),
			rend.hookedRend,
		),
	)
}

func renderError(err error) string {
	return fmt.Sprintf(
		`<div class="markdown-error">
			%s
		</div>`,
		err.Error(),
	)
}

// -------------------------------------------------------------------------- //
// MARKDOWN
// -------------------------------------------------------------------------- //

// Markdown renders the node as markdown string.
func (*NodeRender) Markdown(n *graph.Node) string {
	return parseContent(n, linkFormatAPI)
}

var ratTagRegex = regexp.MustCompile(
	`<rat(\s([^>]+))>`,
)

func parseContent(n *graph.Node, lf linkFormat) string {
	matches := ratTagRegex.FindAllIndex([]byte(n.Content), -1)
	parsedSource := n.Content

	for _, match := range ReverseSlice(matches) {
		tag := n.Content[match[0]:match[1]]

		parsed, err := parseRatTag(n, tag, lf)
		if err != nil {
			log.Warning("failed to parse rat tag", tag, err)
		}

		left := parsedSource[:match[0]]
		right := parsedSource[match[1]:]

		parsedSource = left + parsed + right
	}

	return parsedSource
}

// ReverseSlice reverses a slice.
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
func parseRatTag(n *graph.Node, tag string, lf linkFormat) (string, error) {
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
			return parseRatTagLink(n, args[0], args[1], lf)
		}

		if len(args) == 1 {
			return parseRatTagLink(n, args[0], "", lf)
		}

		return "", errors.New("too many arguments")

	case ratTagKeywordGraph:
		if len(args) == 0 {
			return parseRatTagGraph(n, "-1", lf)
		}

		return parseRatTagGraph(n, args[0], lf)

	case ratTagKeywordTodo:
		return parseRatTagTodo(n, args)

	default:
		return "", errors.New("unknown keyword")
	}
}

// -------------------------------------------------------------------------- //
// link
// -------------------------------------------------------------------------- //

func parseRatTagLink(
	n *graph.Node, linkID, name string, lf linkFormat,
) (string, error) {
	id, err := uuid.FromString(linkID)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse link id")
	}

	node, err := n.Store.GetByID(id)
	if err != nil {
		return "", errors.Wrap(err, "failed to get node by ID")
	}

	return mdLink(node, name, lf), nil
}

func mdLink(n *graph.Node, name string, lf linkFormat) string {
	if name == "" {
		name = n.Name
	}

	return fmt.Sprintf("[%s](%s)", name, link(n.Path, lf))
}

func link(path string, lf linkFormat) string {
	switch lf {
	case linkFormatWeb:
		var (
			u url.URL
			q = make(url.Values)
		)

		u.Path = "/view/"

		q.Add("node", path)

		u.RawQuery = q.Encode()

		return u.String()
	case linkFormatAPI:
		l, err := url.JoinPath("/nodes/", path)
		if err != nil {
			log.Error("failed to join url path", err)
		}

		return l
	default:
		log.Error("unknown link format", lf)

		return ""
	}
}

// -------------------------------------------------------------------------- //
// graph
// -------------------------------------------------------------------------- //

func parseRatTagGraph(
	n *graph.Node, depth string, lf linkFormat,
) (string, error) {
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
					strings.Repeat("\t", depth), mdLink(node, "", lf),
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
// todo
// -------------------------------------------------------------------------- //

var todoRegex = regexp.MustCompile(
	"```todo\n((?:.*\n)*?)```",
)

//nolint:gocognit,gocyclo,cyclop
func parseRatTagTodo(
	n *graph.Node, args []string,
) (string, error) {
	var (
		err            error
		parent         *graph.Node
		depth          = math.MaxInt
		filterPriority bool
		argIdx         int
	)

	for argIdx < len(args) {
		err := func() error {
			// consumes one
			defer func() { argIdx++ }()

			arg := args[argIdx]

			if arg == "filter" {
				filterPriority = true

				return nil
			}

			// can consume two
			if argIdx+1 >= len(args) {
				return nil
			}

			// consumes two
			defer func() { argIdx++ }()

			if arg == "parent" {
				parentID, err := uuid.FromString(args[argIdx+1])
				if err != nil {
					return errors.Wrap(err, "failed to parse parent id")
				}

				parent, err = n.Store.GetByID(parentID)
				if err != nil {
					return errors.Wrap(err, "failed to get parent node")
				}

				return nil
			}

			if arg == "depth" {
				depth, err = strconv.Atoi(args[argIdx+1])
				if err != nil {
					return errors.Wrap(err, "failed to parse depth")
				}

				return nil
			}

			return nil
		}()
		if err != nil {
			return "", err
		}
	}

	if parent == nil {
		parent, err = n.Parent()
		if err != nil {
			return "", errors.Wrap(err, "failed to get parent")
		}
	}

	var (
		todoLists []*todo.TodoList
		derr      error
	)

	err = parent.Walk(
		func(i int, leaf *graph.Node) bool {
			if i == depth {
				return false
			}

			matches := todoRegex.FindAllStringSubmatch(leaf.Content, -1)
			for _, match := range matches {
				todoL, err := todo.Parse(match[1])
				if err != nil {
					log.Warning("failed to parse todo", err)
					derr = errors.Wrap(err, "failed to parse todo")

					return false
				}

				notDone := todoL.NotDone()
				if notDone.Empty() {
					continue
				}

				if filterPriority && !notDone.HasPriority() {
					continue
				}

				notDone.SourceNodePath = leaf.Path
				todoLists = append(todoLists, notDone)
			}

			return true
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to walk graph")
	}

	if derr != nil {
		return derr.Error(), derr
	}

	sort.SliceStable(
		todoLists,
		func(iIdx, jIdx int) bool {
			i := todoLists[iIdx]
			j := todoLists[jIdx]

			if !i.HasPriority() && !j.HasPriority() {
				return false
			}

			if i.HasPriority() && !j.HasPriority() {
				return true
			}

			if !i.HasPriority() && j.HasPriority() {
				return false
			}

			return i.Priority() > j.Priority()
		},
	)

	rawTodoLists := make([]string, 0, len(todoLists))

	for _, todoList := range todoLists {
		rawTodoLists = append(
			rawTodoLists,
			todoList.Markdown(),
		)
	}

	return strings.Join(rawTodoLists, "\n"), nil
}
