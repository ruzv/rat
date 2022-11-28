package todo

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// var log = logging.MustGetLogger("render.todo")

// Parse parses a todo list from a markdown string.
//
//nolint:gocognit,gocyclo,cyclop
func Parse(raw string) (*TodoList, error) {
	lines := strings.Split(raw, "\n")

	//nolint:prealloc
	var (
		first          = true
		todoLines      []string
		rawTodos       []string
		due            time.Time
		size           time.Duration
		sourceNodePath string
		err            error
	)

	for _, line := range lines {
		if strings.HasPrefix(line, "due=") {
			line = strings.TrimPrefix(line, "due=")
			if len(line) == 0 {
				continue
			}

			due, err = parseTime(line)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse due date")
			}

			continue
		}

		if strings.HasPrefix(line, "size=") {
			line = strings.TrimPrefix(line, "size=")
			if len(line) == 0 {
				continue
			}

			size, err = time.ParseDuration(line)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse size")
			}

			continue
		}

		if strings.HasPrefix(line, "sourceNodePath=") {
			sourceNodePath = strings.TrimPrefix(line, "sourceNodePath=")

			continue
		}

		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "x ") {
			if first {
				first = false
			} else {
				rawTodos = append(rawTodos, strings.Join(todoLines, "\n"))
				todoLines = []string{}
			}
		}

		todoLines = append(todoLines, line)
	}

	if len(todoLines) != 0 {
		rawTodos = append(rawTodos, strings.Join(todoLines, "\n"))
	}

	todos := make([]Todo, 0, len(rawTodos))

	for _, rawTodo := range rawTodos {
		var done bool

		rawTodo = strings.TrimSpace(rawTodo)

		if len(rawTodo) == 0 {
			continue
		}

		if strings.HasPrefix(rawTodo, "x ") {
			done = true
		}

		todos = append(todos, Todo{
			Done: done,
			Text: rawTodo[2:],
		})
	}

	return &TodoList{
		List:           todos,
		Due:            due,
		Size:           size,
		SourceNodePath: sourceNodePath,
	}, nil
}

func parseTime(raw string) (time.Time, error) {
	t, err := time.Parse("02.01.2006", raw)
	if err == nil {
		return t, nil
	}

	t, err = time.Parse("02.01.2006.", raw)
	if err == nil {
		return t, nil
	}

	t, err = time.Parse("2.01.2006", raw)
	if err == nil {
		return t, nil
	}

	t, err = time.Parse("2.01.2006.", raw)
	if err == nil {
		return t, nil
	}

	t, err = time.Parse("2.01", raw)
	if err == nil {
		return t.AddDate(time.Now().Year(), 0, 0), nil
	}

	t, err = time.Parse("2.01.", raw)
	if err == nil {
		return t.AddDate(time.Now().Year(), 0, 0), nil
	}

	return time.Time{}, errors.New("failed to parse time in any format")
}

// TodoList describes a single todo list.
type TodoList struct {
	List           []Todo
	Due            time.Time
	Size           time.Duration
	SourceNodePath string
}

// CopyMeta returns a new todo list with the same meta data as the current one.
func (tl *TodoList) CopyMeta() *TodoList {
	return &TodoList{
		Due:            tl.Due,
		Size:           tl.Size,
		SourceNodePath: tl.SourceNodePath,
	}
}

// Done returns a net todo list with only the done todos.
func (tl *TodoList) Done() *TodoList {
	var done []Todo

	for _, td := range tl.List {
		if td.Done {
			done = append(done, td)
		}
	}

	ret := tl.CopyMeta()
	ret.List = done

	return ret
}

// NotDone returns a new todo list with only the not done todos.
func (tl *TodoList) NotDone() *TodoList {
	var notDone []Todo

	for _, td := range tl.List {
		if !td.Done {
			notDone = append(notDone, td)
		}
	}

	ret := tl.CopyMeta()
	ret.List = notDone

	return ret
}

// Markdown returns the todo list as a markdown string.
func (tl *TodoList) Markdown() string {
	lines := make([]string, 0, len(tl.List)+5)

	lines = append(
		lines,
		"```todo",
		tl.DueString(),
		tl.SizeString(),
		fmt.Sprintf("sourceNodePath=%s", tl.SourceNodePath),
	)

	for _, td := range tl.List {
		lines = append(lines, td.Markdown())
	}

	lines = append(lines, "```")

	return strings.Join(lines, "\n")
}

// DueString returns the due date as a string in the format "due=02.01.2006".
func (tl *TodoList) DueString() string {
	if !tl.HasDue() {
		return ""
	}

	return fmt.Sprintf("due=%s", tl.Due.Format("02.01.2006"))
}

// SizeString returns the size as a string in the format "size=1h30m".
func (tl *TodoList) SizeString() string {
	if !tl.HasSize() {
		return ""
	}

	return fmt.Sprintf("size=%s", tl.Size.String())
}

// PriorityString returns the priority as a string in the format "priority=1".
func (tl *TodoList) PriorityString() string {
	if !tl.HasPriority() {
		return ""
	}

	return fmt.Sprintf("priority=%.2f", tl.Priority())
}

// Empty returns true if the todo list is empty.
func (tl *TodoList) Empty() bool {
	return len(tl.List) == 0
}

// HasDue returns true if the todo list has a due date.
func (tl *TodoList) HasDue() bool {
	return !tl.Due.IsZero()
}

// HasSize returns true if the todo list has a size.
func (tl *TodoList) HasSize() bool {
	return tl.Size != 0
}

// HasPriority returns true if the todo list has a priority.
func (tl *TodoList) HasPriority() bool {
	return tl.HasDue() && tl.HasSize()
}

var timeNow = time.Now

// Priority calculates the todo lists priority, the higher the value the higher
// the priority.
func (tl *TodoList) Priority() float64 {
	x := timeNow().Unix()
	d := tl.Due.Unix()

	diff := float64(d - x) // (60 * 60 * 24)
	s := tl.Size.Seconds() // (60 * 60 * 24)

	if diff <= 0 {
		return math.Inf(+1)
	}

	// s / diff - as long as diff is not 0

	return s / diff
}

// Todo describes a single todo. Multiple todos can be grouped in a todo list.
type Todo struct {
	Done bool
	Text string
}

// Markdown returns the todo as a markdown string.
func (td Todo) Markdown() string {
	i := "-"
	if td.Done {
		i = "x"
	}

	return fmt.Sprintf("%s %s", i, td.Text)
}
