package render

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"strings"

	"private/rat/graph"
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
	ts   *TemplateStore
	p    graph.Provider
	rend *html.Renderer
}

// NewRenderer creates a new Renderer.
func NewRenderer(ts *TemplateStore, p graph.Provider) *Renderer {
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
	return string(
		markdown.ToHTML(
			[]byte(token.TransformContentTokens(n, r.p)),
			parser.NewWithExtensions(parser.CommonExtensions),
			r.rend,
		),
	)
}

// TemplateStore contains all templates used by the renderer.
type TemplateStore struct {
	link      *template.Template
	codeBlock *template.Template
}

// LinkTemplData contains the data used to render a link.
type LinkTemplData struct {
	Link string
	Name string
}

// Link renders a link.
func (ts *TemplateStore) Link(w io.Writer, data LinkTemplData) error {
	err := ts.link.Execute(w, data)
	if err != nil {
		return errors.Wrap(err, "failed to execute link template")
	}

	return nil
}

// CodeBlock renders a code block.
func (ts *TemplateStore) CodeBlock(w io.Writer, lines []string) error {
	err := ts.codeBlock.Execute(w, struct{ Lines []string }{lines})
	if err != nil {
		return errors.Wrap(err, "failed to execute link template")
	}

	return nil
}

// FileTemplateStore creates a new TemplateStore with the templates from the
// specified directory.
func FileTemplateStore(templateFS fs.FS) (*TemplateStore, error) {
	ts := &TemplateStore{
		link:      &template.Template{},
		codeBlock: &template.Template{},
	}

	for name, dest := range map[string]*template.Template{
		"link.tmpl":      ts.link,
		"codeBlock.tmpl": ts.codeBlock,
	} {
		templ, err := template.ParseFS(templateFS, name)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s template", name)
		}

		*dest = *templ
	}

	return ts, nil
}

func (r *Renderer) hook() html.RenderNodeFunc {
	return func(
		w io.Writer, node ast.Node, entering bool,
	) (ast.WalkStatus, bool) {
		var parsed string

		switch n := node.(type) {
		case *ast.CodeBlock:
			if string(n.Info) != "todo" {
				return r.renderCodeBlock(w, string(n.Literal))
			}

			parsed = renderTodoList(r.rend, string(n.Literal))
		case *ast.Code:
			parsed = fmt.Sprintf(
				"<span class=\"markdown-code\">%s</span>", string(n.Literal),
			)
		case *ast.Link:
			return r.renderLink(w, n, entering)
		default:
			return ast.GoToNext, false // false - didn't enter current ast.Node
		}

		_, err := io.WriteString(w, parsed)
		if err != nil {
			return ast.GoToNext, false
		}

		return ast.GoToNext, true
	}
}

func (r *Renderer) renderLink(
	w io.Writer, n *ast.Link, entering bool,
) (ast.WalkStatus, bool) {
	if len(n.Children) != 1 {
		// unknown structure, let gomarkdown handle it
		return ast.GoToNext, false
	}

	// skip exiting only for structures we know how to handle
	if !entering {
		return ast.GoToNext, false
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
		LinkTemplData{
			Link: link,
			Name: name,
		},
	)
	if err != nil {
		log.Error("failed to render link", err)

		return ast.GoToNext, false
	}

	return ast.SkipChildren, true
}

func (r *Renderer) renderCodeBlock(
	w io.Writer, raw string,
) (ast.WalkStatus, bool) {
	lines := strings.Split(raw, "\n")
	if len(lines) != 0 {
		lines = lines[:len(lines)-1] // last line is always empty
	}

	err := r.ts.CodeBlock(w, lines)
	if err != nil {
		log.Error("failed to render code block", err)

		return ast.GoToNext, false
	}

	return ast.GoToNext, true
}

func renderTodoList(rend *html.Renderer, raw string) string {
	todoL, err := todo.Parse(raw)
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
				util.Link(todoL.SourceNodePath, string(todoL.SourceNodePath)),
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
			<div class="markdown-todo-checkbox-border">
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
