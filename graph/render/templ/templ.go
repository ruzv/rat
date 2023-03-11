package templ

import (
	"io"
	"io/fs"
	"text/template"

	"github.com/pkg/errors"
)

// TemplateStore contains all templates used by the renderer.
type TemplateStore struct {
	link      *template.Template
	codeBlock *template.Template
	code      *template.Template
	todo      *template.Template
}

// FileTemplateStore creates a new templateStore with the templates from the
// specified directory.
func FileTemplateStore(templateFS fs.FS) (*TemplateStore, error) {
	ts := &TemplateStore{
		link:      &template.Template{},
		codeBlock: &template.Template{},
		code:      &template.Template{},
		todo:      &template.Template{},
	}

	for name, dest := range map[string]*template.Template{
		"link.tmpl":      ts.link,
		"codeBlock.tmpl": ts.codeBlock,
		"code.tmpl":      ts.code,
		"todo.tmpl":      ts.todo,
	} {
		templ, err := template.ParseFS(templateFS, name)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s template", name)
		}

		*dest = *templ
	}

	return ts, nil
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
		return errors.Wrap(err, "failed to execute code block template")
	}

	return nil
}

// Code renders a `code` element.
func (ts *TemplateStore) Code(w io.Writer, text string) error {
	err := ts.code.Execute(w, struct{ Text string }{Text: text})
	if err != nil {
		return errors.Wrap(err, "failed to execute code template")
	}

	return nil
}

// TodoEntryTemplData contains the data used to render a todo entry.
type TodoEntryTemplData struct {
	Done    bool
	Content string
}

// Todo renders a list of todo entires.
func (ts *TemplateStore) Todo(
	w io.Writer,
	entires []TodoEntryTemplData,
	hints map[string]string,
) error {
	err := ts.todo.Execute(
		w,
		struct {
			Hints   map[string]string
			Entries []TodoEntryTemplData
		}{
			Hints:   hints,
			Entries: entires,
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to execute todo entry template")
	}

	return nil
}
