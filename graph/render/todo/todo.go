package todo

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"private/rat/graph"
	"private/rat/graph/util"

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

// TODO: this will be used when parsing todo metadata.
//
//nolint:unused
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
