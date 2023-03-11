package todo

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"private/rat/graph"
	"private/rat/graph/util"

	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var todoRe = regexp.MustCompile(
	"\x60\x60\x60todo\n((?:(?:.*)\n)*?)\x60\x60\x60",
)

var log = logging.MustGetLogger("graph.render.todo")

// HintType describes todo hint types.
type HintType string

const (
	// None is a hint type that does not provide any hints.
	None HintType = "none"
	// Due is a hint type that provides a due date.
	Due HintType = "due"
	// Size is a hint type that provides a size.
	Size HintType = "size"
	// Src is a hint type that shows the source node path of todo.
	Src HintType = "src"
)

// Todo represents a todo list.
type Todo struct {
	Entries []*TodoEntry
	Hints   map[HintType]interface{}
}

// TODO: FIGURE OUT HOW TO HANDLE HINTS, MYb just an interface, and sort,
// contains, ... handling is done by casting
// due
// size
// tags
// sort, compare
// filter - contains tag

// ParseNode parses a todo lists from a node.
func ParseNode(n *graph.Node) ([]*Todo, error) {
	var todos []*Todo //nolint:prealloc

	matches := todoRe.FindAllStringSubmatch(n.Content, -1)
	for _, match := range matches {
		t, err := Parse(match[1])
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse todo")
		}

		t.Hints[Src] = n.Path

		todos = append(todos, t)
	}

	return todos, nil
}

// Parse parses a todo list from a raw string.
func Parse(raw string) (*Todo, error) {
	lf := &lineFeed{
		// filter empty
		lines: util.Filter(
			strings.Split(raw, "\n"),
			func(s string) bool { return len(s) > 0 },
		),
	}

	// find todo metadata

	var (
		entries []*TodoEntry
		hints   = make(map[HintType]interface{})
	)

	for lf.next() {
		line := lf.peek()

		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "x ") {
			e, err := parseEntry(
				append(
					[]string{lf.pop()},
					lf.popUntil(
						func(s string) bool {
							return strings.HasPrefix(s, "- ") ||
								strings.HasPrefix(s, "x ")
						},
					)...,
				),
			)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse entry")
			}

			entries = append(entries, e)

			continue
		}

		if strings.Contains(line, "=") {
			hType, hValue, err := parseHint(lf.pop())
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse hint")
			}

			hints[hType] = hValue

			continue
		}

		log.Warningf("unknown line - %q", lf.pop())
	}

	return &Todo{
		Entries: entries,
		Hints:   hints,
	}, nil
}

func parseHint(line string) (HintType, interface{}, error) {
	parts := strings.Split(line, "=")

	if len(parts) != 2 {
		return None, nil, errors.Errorf("invalid hint - %q", line)
	}

	hType := strings.TrimSpace(parts[0])
	hValue := strings.TrimSpace(parts[1])

	switch HintType(hType) {
	case Due:
		due, err := parseTime(hValue)
		if err != nil {
			return None, nil, errors.Wrap(err, "failed to parse due date")
		}

		return Due, due, nil
	case Size:
		size, err := time.ParseDuration(hValue)
		if err != nil {
			return None, nil, errors.Wrap(err, "failed to parse size")
		}

		return Size, size, nil

	case Src:
		return Src, hValue, nil
	default:
		log.Warningf("unknown hint type - %q, with value - %q", hType, hValue)

		return None, nil, nil // unknown hint type, not and error, just ignore.
	}
}

// Markdown returns the markdown representation of the todo list.
func (t *Todo) Markdown() string {
	hints := make([]string, 0, len(t.Hints))

	for k, v := range t.StringHints() {
		hints = append(hints, fmt.Sprintf("%s = %s", k, v))
	}

	return fmt.Sprintf(
		"```todo\n%s\n%s\n```",
		strings.Join(hints, "\n"),
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

// StringHints converts the todos hins to a string map, that can be used to
// render the todos template.
func (t *Todo) StringHints() map[string]string {
	hints := make(map[string]string)

	for hType, hValue := range t.Hints {
		if hValue == nil {
			continue
		}

		strValue := func() string {
			switch hType {
			case Due:
				v, ok := hValue.(time.Time)
				if !ok {
					return fmt.Sprintf("%v", hValue)
				}

				return v.Format("02.01.2006")
			case Size:
				v, ok := hValue.(time.Duration)
				if !ok {
					return fmt.Sprintf("%v", hValue)
				}

				return v.String()
			default:
				return fmt.Sprintf("%v", hValue)
			}
		}()

		hints[string(hType)] = strValue
	}

	return hints
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

	return time.Time{},
		errors.Errorf("failed to parse %q as time in any format", raw)
}

// ByDue produces the sort arguments for sorting todos by due date.
func ByDue(s []*Todo) ([]*Todo, func(i, j int) bool) {
	return s, func(i, j int) bool {
		iT := s[i].hintDue()
		jT := s[j].hintDue()

		if iT == nil && jT == nil {
			return false
		}

		if iT == nil {
			return true
		}

		if jT == nil {
			return false
		}

		return iT.Before(*jT)
	}
}

func (t *Todo) hintDue() *time.Time {
	due, ok := t.Hints[Due]
	if !ok {
		return nil
	}

	tDue, ok := due.(time.Time)
	if !ok {
		return nil
	}

	return &tDue
}

// Done returns true if all entries are done.
func (t *Todo) Done() bool {
	for _, e := range t.Entries {
		if !e.Done {
			return false
		}
	}

	return true
}
