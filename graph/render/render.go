package render

import (
	"io"
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
				return r.renderCode(w, string(n.Literal))
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

	id, err := uuid.FromString(link)
	if err == nil {
		n, err := r.p.GetByID(id)
		if err != nil {
			return ast.GoToNext, false, errors.Wrap(err, "failed to get node")
		}

		link, err = pathutil.URL(n.Path)
		if err != nil {
			return ast.GoToNext,
				false,
				errors.Wrap(err, "failed to get node path")
		}

		if len(strings.TrimSpace(name)) == 0 {
			name = n.Name
		}
	}

	err = r.ts.Link(
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

func (r *Renderer) renderCode(
	w io.Writer, raw string,
) (ast.WalkStatus, bool, error) {
	err := r.ts.Code(w, raw)
	if err != nil {
		return ast.GoToNext, false, errors.Wrap(err, "failed to render code")
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

	err = r.ts.Todo(
		w,
		util.Map(
			t.OrderEntries(),
			func(e *todo.TodoEntry) templ.TodoEntryTemplData {
				return templ.TodoEntryTemplData{
					Done:    e.Done,
					Content: r.render(e.Text),
				}
			},
		),
		util.Map(
			t.OrderHints(),
			func(h *todo.Hint) templ.TodoHintTemplData {
				return templ.TodoHintTemplData{
					Type:  string(h.Type),
					Value: h.HTML(),
				}
			},
		),
	)
	if err != nil {
		return ast.GoToNext, false, errors.Wrap(err, "failed to render todo")
	}

	return ast.GoToNext, true, nil
}
