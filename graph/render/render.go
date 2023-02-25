package render

import (
	"io"
	"sort"
	"strings"

	"private/rat/graph"
	"private/rat/graph/render/templ"
	"private/rat/graph/render/todo"
	"private/rat/graph/token"
	"private/rat/graph/util"
	pathutil "private/rat/graph/util/path"

	"github.com/gofrs/uuid"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var log = logging.MustGetLogger("render")

// Renderer allows rendering nodes markdown content to HTML.
type Renderer struct {
	ts   *templ.TemplateStore
	p    graph.Provider
	rend *html.Renderer
}

// NewRenderer creates a new Renderer.
func NewRenderer(ts *templ.TemplateStore, p graph.Provider) *Renderer {
	r := &Renderer{
		ts:   ts,
		p:    p,
		rend: &html.Renderer{}, // allocate, temp value
	}

	*r.rend = *html.NewRenderer(html.RendererOptions{RenderNodeHook: r.hook()})

	return r
}

// Render parses nodes content, converts tokens into markdown and renders it to
// HTML.
func (r *Renderer) Render(n *graph.Node) string {
	return r.render(token.TransformContentTokens(n, r.p))
}

func (r *Renderer) render(raw string) string {
	return string(
		markdown.ToHTML(
			[]byte(raw),
			parser.NewWithExtensions(parser.CommonExtensions),
			r.rend,
		),
	)
}

func (r *Renderer) hook() html.RenderNodeFunc {
	return func(
		w io.Writer, node ast.Node, entering bool,
	) (ast.WalkStatus, bool) {
		ws, e, err := func() (ast.WalkStatus, bool, error) {
			switch n := node.(type) {
			case *ast.CodeBlock:
				if string(n.Info) == "todo" {
					return r.renderTodo(w, string(n.Literal))
				}

				return r.renderCodeBlock(w, string(n.Literal))
			case *ast.Code:
				err := r.ts.Code(w, string(n.Literal))
				if err != nil {
					return ast.GoToNext,
						false,
						errors.Wrap(err, "failed to render code")
				}

				return ast.GoToNext, true, nil
			case *ast.Link:
				return r.renderLink(w, n, entering)
			default:
				// false - didn't enter current ast.Node
				return ast.GoToNext, false, nil
			}
		}()
		if err != nil {
			log.Error("failed to render node", err)
		}

		return ws, e
	}
}

func (r *Renderer) renderLink(
	w io.Writer, n *ast.Link, entering bool,
) (ast.WalkStatus, bool, error) {
	if len(n.Children) != 1 {
		// unknown structure, let gomarkdown handle it
		return ast.GoToNext, false, errors.New("unknown link structure")
	}

	// skip exiting only for structures we know how to handle
	if !entering {
		return ast.GoToNext, false, nil
	}

	link := string(n.Destination)
	name := string(n.Children[0].AsLeaf().Literal)

	func() {
		id, err := uuid.FromString(link)
		if err != nil {
			return
		}

		n, err := r.p.GetByID(id)
		if err != nil {
			return
		}

		link = pathutil.URL(n.Path)

		if len(strings.TrimSpace(name)) != 0 {
			return
		}

		name = n.Name
	}()

	err := r.ts.Link(
		w,
		templ.LinkTemplData{
			Link: link,
			Name: name,
		},
	)
	if err != nil {
		return ast.GoToNext, false, errors.Wrap(err, "failed to render link")
	}

	return ast.SkipChildren, true, nil
}

func (r *Renderer) renderCodeBlock(
	w io.Writer, raw string,
) (ast.WalkStatus, bool, error) {
	lines := strings.Split(raw, "\n")
	if len(lines) != 0 {
		lines = lines[:len(lines)-1] // last line is always empty
	}

	err := r.ts.CodeBlock(w, lines)
	if err != nil {
		return ast.GoToNext,
			false,
			errors.Wrap(err, "failed to render code block")
	}

	return ast.GoToNext, true, nil
}

func (r *Renderer) renderTodo(
	w io.Writer, raw string,
) (ast.WalkStatus, bool, error) {
	t, err := todo.Parse(raw)
	if err != nil {
		return ast.GoToNext, false, errors.Wrap(err, "failed to parse todo")
	}

	sort.Sort(t)

	err = r.ts.Todo(
		w,
		util.Map(
			t.Entries,
			func(e *todo.TodoEntry) templ.TodoEntryTemplData {
				return templ.TodoEntryTemplData{
					Done:    e.Done,
					Content: r.render(e.Text),
				}
			},
		),
	)
	if err != nil {
		return ast.GoToNext, false, errors.Wrap(err, "failed to render todo")
	}

	return ast.GoToNext, true, nil
}

// func renderTodoList(rend *html.Renderer, raw string) string {
// 	todoL, err := todo.Parse(raw)
// 	if err != nil {
// 		log.Warning("failed to parse todo list", err)

// 		return renderError(errors.Wrap(err, "failed to parse todo"))
// 	}

// 	if todoL.Empty() {
// 		return ""
// 	}

// 	list := append(todoL.NotDone().List, todoL.Done().List...) //nolint:gocritic
// 	parts := make([]string, 0, len(list))

// 	for _, todo := range list {
// 		parts = append(parts, renderTodo(rend, todo))
// 	}

// 	var header []string
// 	header = append(header, `<div class="markdown-todo-params">`)

// 	if len(todoL.SourceNodePath) != 0 {
// 		header = append(
// 			header,
// 			fmt.Sprintf(
// 				`<span class="markdown-code"><a href="%s">%s</a></span>`,
// 				util.Link(todoL.SourceNodePath, string(todoL.SourceNodePath)),
// 				todoL.SourceNodePath,
// 			),
// 		)
// 	}

// 	if len(todoL.PriorityString()) != 0 {
// 		header = append(
// 			header,
// 			fmt.Sprintf(
// 				`<div><span class="markdown-code">%s</span></div>`,
// 				todoL.PriorityString(),
// 			),
// 		)
// 	}

// 	if len(todoL.DueString()) != 0 {
// 		header = append(
// 			header,
// 			fmt.Sprintf(
// 				`<div><span class="markdown-code">%s</span></div>`,
// 				todoL.DueString(),
// 			),
// 		)
// 	}

// 	if len(todoL.SizeString()) != 0 {
// 		header = append(
// 			header,
// 			fmt.Sprintf(
// 				`<div><span class="markdown-code">%s</span></div>`,
// 				todoL.SizeString(),
// 			),
// 		)
// 	}

// 	header = append(header, `</div>`)

// 	return fmt.Sprintf(
// 		`<div class="markdown-todo-list">
// 			<div class="markdown-todo-header">
// 				%s
// 			</div>
// 			%s
// 		</div>`,
// 		strings.Join(header, "\n"),
// 		strings.Join(parts, "\n"),
// 	)
// }

// func renderTodo(rend *html.Renderer, td todo.Todo) string {
// 	var checked string

// 	if td.Done {
// 		checked = " markdown-todo-checkbox-check-checked"
// 	}

// 	return fmt.Sprintf(
// 		`<div class="markdown-todo">
// 			<div class="markdown-todo-checkbox-border">
// 				<div class="markdown-todo-checkbox-check%s">
// 				</div>
// 			</div>
// 			<div class="markdown-todo-text">
// 				%s
// 			</div>
// 		</div>`,
// 		checked,
// 		markdown.ToHTML(
// 			[]byte(td.Text),
// 			parser.NewWithExtensions(parser.CommonExtensions),
// 			rend,
// 		),
// 	)
// }

// func renderError(err error) string {
// 	return fmt.Sprintf(
// 		`<div class="markdown-error">
// 			%s
// 		</div>`,
// 		err.Error(),
// 	)
// }
