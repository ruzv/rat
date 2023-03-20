package todo

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

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
