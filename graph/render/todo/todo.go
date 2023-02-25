package todo

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"private/rat/graph"
	"private/rat/graph/util"
	pathutil "private/rat/graph/util/path"

	"github.com/pkg/errors"
)

var todoRe = regexp.MustCompile("```todo\n([[:ascii:]]*?)```")

// var log = logging.MustGetLogger("render.todo")

// Todo represents a todo list.
type Todo struct {
	Entries []*TodoEntry
}

// ParseNode parses a todo lists from a node.
func ParseNode(n *graph.Node) ([]*Todo, error) {
	var todos []*Todo //nolint:prealloc

	matches := todoRe.FindAllStringSubmatch(n.Content, -1)
	for _, match := range matches {
		t, err := Parse(match[1])
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse todo")
		}

		todos = append(todos, t)
	}

	return todos, nil
}

// Parse parses a todo list from a raw string.
func Parse(raw string) (*Todo, error) {
	// filter empty
	lines := util.Filter(
		strings.Split(raw, "\n"),
		func(s string) bool { return len(s) > 0 },
	)

	//nolint:prealloc
	var (
		entries    []*TodoEntry
		entryLines []string
	)

	for _, line := range lines {
		if (strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "x ")) &&
			len(entryLines) > 0 {
			e, err := parseEntry(entryLines)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse entry")
			}

			entries = append(entries, e)

			entryLines = []string{}
		}

		entryLines = append(entryLines, line)
	}

	e, err := parseEntry(entryLines)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse entry")
	}

	entries = append(entries, e)

	return &Todo{
		Entries: entries,
	}, nil
}

// Markdown returns the markdown representation of the todo list.
func (t *Todo) Markdown() string {
	return fmt.Sprintf(
		"```todo\n%s\n```",
		strings.Join(
			util.Map(
				t.Entries,
				func(e *TodoEntry) string {
					return e.markdown()
				},
			),
			"\n",
		),
	)
}

var _ sort.Interface = (*Todo)(nil)

// Len is the number of elements in the collection.
func (t *Todo) Len() int {
	return len(t.Entries)
}

// Less reports whether the element with index i
// must sort before the element with index j.
//
// If both Less(i, j) and Less(j, i) are false,
// then the elements at index i and j are considered equal.
// Sort may place equal elements in any order in the final result,
// while Stable preserves the original input order of equal elements.
//
// Less must describe a transitive ordering:
//   - if both Less(i, j) and Less(j, k) are true,
//     then Less(i, k) must be true as well.
//   - if both Less(i, j) and Less(j, k) are false,
//     then Less(i, k) must be false as well.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
// See Float64Slice.Less for a correct implementation for floating-point values.
func (t *Todo) Less(i, j int) bool {
	if t.Entries[i].Done && !t.Entries[j].Done {
		return false
	}

	return true
}

// Swap swaps the elements with indexes i and j.
func (t *Todo) Swap(i, j int) {
	t.Entries[i], t.Entries[j] = t.Entries[j], t.Entries[i]
}

// Empty returns true if the todo list is empty.
func (t *Todo) Empty() bool {
	if t == nil {
		return true
	}

	return len(t.Entries) == 0
}

// TodoEntry represents a single todo entry.
type TodoEntry struct {
	Done bool
	Text string
}

// parseEntry parses a single todo entry from a list of lines.
func parseEntry(lines []string) (*TodoEntry, error) {
	text := strings.Join(lines, "\n")

	if len(text) == 0 {
		return nil, errors.New("empty todo entry")
	}

	var done bool

	switch text[0] {
	case '-':
		done = false
	case 'x':
		done = true
	default:
		return nil, errors.Errorf("invalid todo entry - %q", text)
	}

	return &TodoEntry{
		Done: done,
		Text: text[1:],
	}, nil
}

func (te *TodoEntry) markdown() string {
	if te.Done {
		return fmt.Sprintf("x %s", te.Text)
	}

	return fmt.Sprintf("- %s", te.Text)
}

// Parse parses a todo list from a markdown string.
//
//nolint:gocognit,gocyclo,cyclop
// func Parse(raw string) (*TodoList, error) {
// 	lines := strings.Split(raw, "\n")

// 	//nolint:prealloc
// 	var (
// 		first          = true
// 		todoLines      []string
// 		rawTodos       []string
// 		due            time.Time
// 		size           time.Duration
// 		sourceNodePath pathutil.NodePath
// 		err            error
// 	)

// 	for _, line := range lines {
// 		if strings.HasPrefix(line, "due=") {
// 			line = strings.TrimPrefix(line, "due=")
// 			if len(line) == 0 {
// 				continue
// 			}

// 			due, err = parseTime(line)
// 			if err != nil {
// 				return nil, errors.Wrap(err, "failed to parse due date")
// 			}

// 			continue
// 		}

// 		if strings.HasPrefix(line, "size=") {
// 			line = strings.TrimPrefix(line, "size=")
// 			if len(line) == 0 {
// 				continue
// 			}

// 			size, err = time.ParseDuration(line)
// 			if err != nil {
// 				return nil, errors.Wrap(err, "failed to parse size")
// 			}

// 			continue
// 		}

// 		if strings.HasPrefix(line, "sourceNodePath=") {
// 			sourceNodePath = pathutil.NodePath(
// 				strings.TrimPrefix(line, "sourceNodePath="),
// 			)

// 			continue
// 		}

// 		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "x ") {
// 			if first {
// 				first = false
// 			} else {
// 				rawTodos = append(rawTodos, strings.Join(todoLines, "\n"))
// 				todoLines = []string{}
// 			}
// 		}

// 		todoLines = append(todoLines, line)
// 	}

// 	if len(todoLines) != 0 {
// 		rawTodos = append(rawTodos, strings.Join(todoLines, "\n"))
// 	}

// 	todos := make([]Todo, 0, len(rawTodos))

// 	for _, rawTodo := range rawTodos {
// 		var done bool

// 		rawTodo = strings.TrimSpace(rawTodo)

// 		if len(rawTodo) == 0 {
// 			continue
// 		}

// 		if strings.HasPrefix(rawTodo, "x ") {
// 			done = true
// 		}

// 		todos = append(todos, Todo{
// 			Done: done,
// 			Text: rawTodo[2:],
// 		})
// 	}

// 	return &TodoList{
// 		List:           todos,
// 		Due:            due,
// 		Size:           size,
// 		SourceNodePath: sourceNodePath,
// 	}, nil
// }

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
	SourceNodePath pathutil.NodePath
}

// CopyMeta returns a new todo list with the same meta data as the current one.
func (tl *TodoList) CopyMeta() *TodoList {
	return &TodoList{
		Due:            tl.Due,
		Size:           tl.Size,
		SourceNodePath: tl.SourceNodePath,
	}
}

// // Done returns a net todo list with only the done todos.
// func (tl *TodoList) Done() *TodoList {
// 	var done []Todo

// 	for _, td := range tl.List {
// 		if td.Done {
// 			done = append(done, td)
// 		}
// 	}

// 	ret := tl.CopyMeta()
// 	ret.List = done

// 	return ret
// }

// // NotDone returns a new todo list with only the not done todos.
// func (tl *TodoList) NotDone() *TodoList {
// 	var notDone []Todo

// 	for _, td := range tl.List {
// 		if !td.Done {
// 			notDone = append(notDone, td)
// 		}
// 	}

// 	ret := tl.CopyMeta()
// 	ret.List = notDone

// 	return ret
// }

// Markdown returns the todo list as a markdown string.
// func (tl *TodoList) Markdown() string {
// 	lines := make([]string, 0, len(tl.List)+5)

// 	lines = append(
// 		lines,
// 		"```todo",
// 		tl.DueString(),
// 		tl.SizeString(),
// 		fmt.Sprintf("sourceNodePath=%s", tl.SourceNodePath),
// 	)

// 	for _, td := range tl.List {
// 		lines = append(lines, td.Markdown())
// 	}

// 	lines = append(lines, "```")

// 	return strings.Join(lines, "\n")
// }

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
type Todo4 struct {
	Done bool
	Text string
}

// // Markdown returns the todo as a markdown string.
// func (td Todo) Markdown() string {
// 	i := "-"
// 	if td.Done {
// 		i = "x"
// 	}

// 	return fmt.Sprintf("%s %s", i, td.Text)
// }
