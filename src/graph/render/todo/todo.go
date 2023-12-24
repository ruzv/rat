package todo

import (
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/render/jsonast"
	"rat/graph/util"
)

var todoRe = regexp.MustCompile(
	"\x60\x60\x60todo\n((?:(?:.*)\n)*?)\x60\x60\x60",
)

// Todo represents a todo list.
//
//nolint:godox // false positive.
type Todo struct {
	Entries []*Entry
	Hints   []*Hint
}

// ParseNode parses a todo lists from a node.
func ParseNode(n *graph.Node) ([]*Todo, error) {
	var todos []*Todo //nolint:prealloc // unknown size.

	matches := todoRe.FindAllStringSubmatch(n.Content, -1)
	for _, match := range matches {
		t, err := Parse(match[1])
		if err != nil {
			return nil, errors.Wrapf(
				err, "failed to parse todo in node - %q", n.Path,
			)
		}

		t.Hints = append(t.Hints, &Hint{Src, n.Path})

		todos = append(todos, t)
	}

	return todos, nil
}

// Parse parses a todo list from a raw string.
func Parse(raw string) (*Todo, error) {
	sf := util.NewStringFeed(
		util.Filter(
			strings.Split(raw, "\n"),
			func(s string) bool { return len(strings.TrimSpace(s)) > 0 },
		),
	)

	var (
		entries []*Entry
		hints   []*Hint
	)

	for sf.More() {
		line := sf.Peek()

		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "x ") {
			e, err := parseEntry(
				append(
					[]string{sf.Pop()},
					sf.PopUntil(
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
			h, err := parseHint(sf.Pop())
			if err != nil {
				if errors.Is(err, errUnknownHint) {
					continue
				}

				return nil, errors.Wrap(err, "failed to parse hint")
			}

			hints = append(hints, h)

			continue
		}

		return nil, errors.Errorf("invalid todo line - %q", line)
	}

	return &Todo{
		Entries: entries,
		Hints:   hints,
	}, nil
}

// Render renders a todo into a JSON AST representation.
func (t *Todo) Render(
	part *jsonast.AstPart,
	n *graph.Node,
	r jsonast.Renderer,
) {
	todoPart := &jsonast.AstPart{
		Type:       "todo",
		Attributes: jsonast.AstAttributes{"hints": t.Hints},
	}

	for _, e := range t.Entries {
		todoEntryPart := todoPart.AddContainer(
			&jsonast.AstPart{
				Type: "todo_entry",
				Attributes: jsonast.AstAttributes{
					"done": e.Done,
				},
			},
			true,
		)

		r.Render(todoEntryPart, n, e.Text)
	}

	part.AddLeaf(todoPart)
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
func (t *Todo) OrderEntries() []*Entry {
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
	var entries []*Entry

	for _, e := range t.Entries {
		if !e.Done {
			entries = append(entries, e)
		}
	}

	t.Entries = entries
}
