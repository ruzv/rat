package todo

import (
	"strings"

	"private/rat/graph/util"

	"github.com/pkg/errors"
)

// FilterRule represents a single filtering rule of multiple todos.
type FilterRule struct {
	Type HintType
	Has  bool
}

// ParseFilterRule parses a filter rule from a raw string.
func ParseFilterRule(raw string) (*FilterRule, error) {
	has := true

	if strings.HasPrefix(raw, "!") {
		has = false
		raw = strings.TrimPrefix(raw, "!")
	}

	_, ok := hintTypeProcessors[HintType(raw)]
	if !ok {
		return nil, errors.Errorf("unknown hint type - %s", raw)
	}

	return &FilterRule{HintType(raw), has}, nil
}

// Filter performs fltration of a list of todos by a list of
// filter rules. Result is returned.
func Filter(todos []*Todo, rules []*FilterRule) []*Todo {
	for _, r := range rules {
		todos = util.Filter(
			todos,
			func(t *Todo) bool {
				has := t.GetHint(r.Type) != nil

				return has == r.Has
			},
		)
	}

	return todos
}
