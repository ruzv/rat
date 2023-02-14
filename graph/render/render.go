package render

import (
	"fmt"
	"io"
	"strings"

	"private/rat/graph/render/todo"
	"private/rat/graph/util"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var log = logging.MustGetLogger("render")

// NewRender creates a new NodeRender.
func NewRenderer() *html.Renderer {
	rend := &html.Renderer{}

	*rend = *html.NewRenderer(
		html.RendererOptions{
			RenderNodeHook: newRenderHook(rend),
		},
	)

	return rend
}

// -------------------------------------------------------------------------- //
// HTML
// -------------------------------------------------------------------------- //

func newRenderHook(
	rend *html.Renderer,
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

func renderTodoList(rend *html.Renderer, n *ast.CodeBlock) string {
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
				util.Link(todoL.SourceNodePath, todoL.SourceNodePath),
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

func renderTodo(rend *html.Renderer, td todo.Todo) string {
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
			parser.NewWithExtensions(parser.CommonExtensions),
			rend,
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
