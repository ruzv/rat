package render

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"

	"github.com/gomarkdown/markdown"
)

type TodoList struct {
	List []Todo
}

func (tl *TodoList) Done() *TodoList {
	var done []Todo

	for _, td := range tl.List {
		if td.Done {
			done = append(done, td)
		}
	}

	// sort.SliceStable(
	// 	done,
	// 	func(i, j int) bool {
	// 		return strings.Compare(done[i].Text, done[j].Text) < 0
	// 	},
	// )

	return &TodoList{
		List: done,
	}
}

func (tl *TodoList) NotDone() *TodoList {
	var notDone []Todo

	for _, td := range tl.List {
		if !td.Done {
			notDone = append(notDone, td)
		}
	}

	// sort.SliceStable(
	// 	notDone,
	// 	func(i, j int) bool {
	// 		return strings.Compare(notDone[i].Text, notDone[j].Text) < 0
	// 	},
	// )

	return &TodoList{
		List: notDone,
	}
}

func (tl *TodoList) Markdown() string {
	lines := make([]string, 0, len(tl.List)+2)

	lines = append(lines, "```todo")

	for _, td := range tl.List {
		lines = append(lines, td.Markdown())
	}

	lines = append(lines, "```")

	return strings.Join(lines, "\n")
}

type Todo struct {
	Done bool
	Text string
}

func (td Todo) render(rend *NodeRender) string {
	var checked string

	if td.Done {
		checked = "markdown-todo-checkbox-check-checked"
	}

	return fmt.Sprintf(
		`<div class="markdown-todo">
			<div class="markdown-todo-checkbox-border">
				<div class="markdown-todo-checkbox-check %s">
				</div>
			</div>			
			<div class="markdown-todo-text">%s</div>
		</div>`,
		checked,
		markdown.ToHTML(
			[]byte(td.Text),
			rend.parser(),
			rend.hookedRend,
		),
	)
}

func (td Todo) Markdown() string {
	i := "-"
	if td.Done {
		i = "x"
	}

	return fmt.Sprintf("%s %s", i, td.Text)
}

func Parse(raw string) (*TodoList, error) {
	buff := bytes.NewBufferString(raw)

	prev, err := buff.ReadByte()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read first byte")
	}

	var (
		chunk  = []byte{prev}
		chunks []string
	)

	for {
		b, err := buff.ReadByte()
		if err != nil {
			if err != io.EOF {
				return nil, errors.Wrap(err, "failed to read byte")
			}

			break
		}

		if prev == '\n' && (b == '-' || b == 'x') {
			chunks = append(chunks, string(chunk))
			chunk = []byte{}
		}

		prev = b
		chunk = append(chunk, b)
	}

	chunks = append(chunks, string(chunk))

	todoList := make([]Todo, 0, len(chunks))

	for _, c := range chunks {
		fields := strings.Fields(c)
		if len(fields) < 2 {
			continue
		}

		marker := fields[0]
		text := strings.Join(fields[1:], " ")

		todoList = append(
			todoList,
			Todo{
				Done: marker == "x",
				Text: text,
			},
		)
	}

	return &TodoList{
		List: todoList,
	}, nil
}

// func parseTODOs(raw []byte) []todo {
// 	buff := bytes.NewBuffer(raw)

// 	prev, err := buff.ReadByte()
// 	if err != nil {
// 		return nil
// 	}

// 	var (
// 		chunk  = []byte{prev}
// 		chunks []string
// 	)

// 	for {
// 		b, err := buff.ReadByte()
// 		if err != nil {
// 			break
// 		}

// 		if prev == '\n' && (b == '-' || b == 'x') {
// 			chunks = append(chunks, string(chunk))
// 			chunk = []byte{}
// 		}

// 		prev = b
// 		chunk = append(chunk, b)
// 	}

// 	chunks = append(chunks, string(chunk))

// 	todoList := make([]Todo, 0, len(chunks))

// 	for _, c := range chunks {
// 		fields := strings.Fields(c)
// 		if len(fields) < 2 {
// 			continue
// 		}

// 		marker := fields[0]
// 		text := strings.Join(fields[1:], " ")

// 		todoList = append(
// 			todoList,
// 			Todo{
// 				done: marker == "x",
// 				text: text,
// 			},
// 		)
// 	}

// 	return todoList
// }
