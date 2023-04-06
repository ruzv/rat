package util

import (
	"io"

	"github.com/pkg/errors"
)

// StringFeed simple read only stack of strings.
type StringFeed struct {
	feed    []string
	current int
}

// NewStringFeed creates a new StringFeed.
func NewStringFeed(feed []string) *StringFeed {
	return &StringFeed{
		feed:    feed,
		current: 0,
	}
}

// More returns true if there are more string to read.
func (lf *StringFeed) More() bool {
	return lf.current < len(lf.feed)
}

// Pop pops the next string from feed.
func (lf *StringFeed) Pop() string {
	if !lf.More() {
		return ""
	}

	defer func() { lf.current++ }()

	return lf.feed[lf.current]
}

// Peek returns the next string from feed without popping it.
func (lf *StringFeed) Peek() string {
	if !lf.More() {
		return ""
	}

	return lf.feed[lf.current]
}

// PopUntil pops the next strings from feed until cond returns true, returning
// all popped strings.
func (lf *StringFeed) PopUntil(cond func(string) bool) []string {
	var lines []string

	for lf.More() {
		if cond(lf.Peek()) {
			break
		}

		lines = append(lines, lf.Pop())
	}

	return lines
}

// PopParts pops the next len(parts) strings from feed and returns an error if
// they don't match.
func (lf *StringFeed) PopParts(parts ...string) error {
	pos := lf.current

	for _, part := range parts {
		if lf.Pop() != part {
			lf.current = pos

			return errors.Errorf("expected %q, got %q", part, lf.Peek())
		}
	}

	return nil
}

// MustPop is like Pop but returns an error if there are no more strings to
// read.
func (lf *StringFeed) MustPop() (string, error) {
	if !lf.More() {
		return "", errors.Wrap(io.EOF, "no more strins in feed")
	}

	return lf.Pop(), nil
}
