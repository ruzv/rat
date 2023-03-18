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

// Todo represents a todo list.
type Todo struct {
	Entries []*TodoEntry
	Hints   []*Hint
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

		t.Hints = append(t.Hints, &Hint{Src, n.Path})

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

	var (
		entries []*TodoEntry
		hints   []*Hint
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
			h, err := parseHint(lf.pop())
			if err != nil {
				if errors.Is(err, errUnknownHint) {
					log.Warningf("unknown hint - %q: %s", line, err.Error())

					continue
				}

				return nil, errors.Wrap(err, "failed to parse hint")
			}

			hints = append(hints, h)

			continue
		}

		log.Warningf("unknown line - %q", lf.pop())
	}

	return &Todo{
		Entries: entries,
		Hints:   hints,
	}, nil
}

// Markdown returns the markdown representation of the todo list.
func (t *Todo) Markdown() string {
	return fmt.Sprintf(
		"```todo\n%s\n%s\n```",
		strings.Join(
			util.Map(
				t.Hints,
				func(h *Hint) string {
					return h.markdown()
				},
			),
			"\n",
		),
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

// OrderHints sorts t.Hints by a predefined order, and returns it.
func (t *Todo) OrderHints() []*Hint {
	priorities := map[HintType]int{
		Src:      0,
		Due:      1,
		Size:     2,
		Priority: 3,
		Tags:     4,
	}

	sort.SliceStable(
		t.Hints,
		func(i, j int) bool {
			return priorities[t.Hints[i].Type] < priorities[t.Hints[j].Type]
		},
	)

	return t.Hints
}

// OrderEntries sorts t.Entries putting not done entries first, and returns it.
func (t *Todo) OrderEntries() []*TodoEntry {
	sort.SliceStable(
		t.Entries,
		func(i, j int) bool {
			if t.Entries[i].Done && !t.Entries[j].Done {
				return false
			}

			return true
		},
	)

	return t.Entries
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

// GetHint returns the hint of the given type. Returns nil if not found.
func (t *Todo) GetHint(hType HintType) *Hint {
	for _, h := range t.Hints {
		if h.Type == hType {
			return h
		}
	}

	return nil
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

// RemoveDoneEntries removes all done entries from t.Entries.
func (t *Todo) RemoveDoneEntries() {
	var entries []*TodoEntry

	for _, e := range t.Entries {
		if !e.Done {
			entries = append(entries, e)
		}
	}

	t.Entries = entries
}
