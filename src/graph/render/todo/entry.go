package todo

import (
	"strings"

	"github.com/pkg/errors"
	"rat/graph/util"
)

// Entry represents a single todo entry.
type Entry struct {
	Done bool
	Text string
}

// parseEntry parses a single todo entry from a list of lines.
func parseEntry(lines []string) (*Entry, error) {
	var done bool

	switch lines[0][0] {
	case '-':
		done = false
	case 'x':
		done = true
	default:
		return nil, errors.Errorf("invalid todo line - %q", lines[0])
	}

	text := strings.Join(
		util.Map(
			lines,
			func(l string) string {
				if strings.HasPrefix(l, "- ") ||
					strings.HasPrefix(l, "x ") ||
					strings.HasPrefix(l, "  ") {
					return l[2:]
				}

				return l
			},
		),
		"\n",
	)

	if text == "" {
		return nil, errors.New("empty todo entry")
	}

	return &Entry{
		Done: done,
		Text: text,
	}, nil
}
