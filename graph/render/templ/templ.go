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
	index     *template.Template
	errorPage *template.Template
}

// FileTemplateStore creates a new templateStore with the templates from the
// specified directory.
func FileTemplateStore(templateFS fs.FS) (*TemplateStore, error) {
	ts := &TemplateStore{
		link:      &template.Template{},
		codeBlock: &template.Template{},
		code:      &template.Template{},
		todo:      &template.Template{},
		index:     &template.Template{},
		errorPage: &template.Template{},
	}

	for name, dest := range map[string]*template.Template{
		"link.html":      ts.link,
		"codeBlock.html": ts.codeBlock,
		"code.html":      ts.code,
		"todo.html":      ts.todo,
		"index.html":     ts.index,
		"errorPage.html": ts.errorPage,
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

// TodoHintTemplData contains the data used to render a todo hint.
type TodoHintTemplData struct {
	Type  string
	Value string
}

// Todo renders a list of todo entires.
func (ts *TemplateStore) Todo(
	w io.Writer,
	entires []TodoEntryTemplData,
	hints []TodoHintTemplData,
) error {
	err := ts.todo.Execute(
		w,
		struct {
			Hints   []TodoHintTemplData
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

// IndexTemplData contains the data used to render the index page.
type IndexTemplData struct {
	ID      string
	Name    string
	Path    string
	Content string
	Leafs   []*IndexTemplLeafData
}

// IndexTemplLeafData contains the data used to render a leaf in the index page.
type IndexTemplLeafData struct {
	Path    string
	Content string
}

// Index renders the index page.
func (ts *TemplateStore) Index(w io.Writer, data *IndexTemplData) error {
	err := ts.index.Execute(w, data)
	if err != nil {
		return errors.Wrap(err, "failed to execute index template")
	}

	return nil
}

// ErrorPageTemplData contains the data used to render the error page.
type ErrorPageTemplData struct {
	Code    int
	Message string
	Cause   string
}

// ErrorPage renders the error page.
func (ts *TemplateStore) ErrorPage(
	w io.Writer, data *ErrorPageTemplData,
) error {
	err := ts.errorPage.Execute(w, data)
	if err != nil {
		return errors.Wrap(err, "failed to execute error page template")
	}

	return nil
}
