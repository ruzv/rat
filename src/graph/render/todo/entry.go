package todo

import (
	"strings"

	"github.com/pkg/errors"
)

// Entry represents a single todo entry.
type Entry struct {
	Done bool
	Text string
}

// parseEntry parses a single todo entry from a list of lines.
func parseEntry(lines []string) (*Entry, error) {
	text := strings.Join(lines, "\n")

	if text == "" {
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

	return &Entry{
		Done: done,
		Text: text[1:],
	}, nil
}
