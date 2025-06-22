package todo

import (
	"strings"
	"time"

	"github.com/pkg/errors"
)

//nolint:gochecknoglobals // TODO: fix.
var sortRules = map[string]*struct {
	equal func(t1, t2 *Todo) bool
	less  func(t1, t2 *Todo) bool
}{
	"done": {
		equal: func(t1, t2 *Todo) bool {
			return t1.Done() == t2.Done()
		},
		less: func(t1, t2 *Todo) bool {
			return t1.Done() && !t2.Done()
		},
	},
	"due": {
		equal: func(t1, t2 *Todo) bool {
			d1 := Value[time.Time](t1.GetHint(Due))
			d2 := Value[time.Time](t2.GetHint(Due))

			return d1.Equal(d2)
		},
		less: func(t1, t2 *Todo) bool {
			d1 := Value[time.Time](t1.GetHint(Due))
			d2 := Value[time.Time](t2.GetHint(Due))

			return d1.Before(d2)
		},
	},
	"priority": {
		equal: func(t1, t2 *Todo) bool {
			p1 := Value[string](t1.GetHint(Priority))
			p2 := Value[string](t2.GetHint(Priority))

			return p1 == p2
		},
		less: func(t1, t2 *Todo) bool {
			p1 := Value[string](t1.GetHint(Priority))
			p2 := Value[string](t2.GetHint(Priority))

			return p1 < p2
		},
	},
}

// SortRule represents a single rule for ordering multiple todos.
type SortRule struct {
	Key     string
	Reverse bool
}

// ParseSortRule parses a sort rule from a string.
func ParseSortRule(raw string) (*SortRule, error) {
	reverse := false

	if strings.HasPrefix(raw, "-") {
		raw = strings.TrimPrefix(raw, "-")
		reverse = true
	}

	_, ok := sortRules[raw]
	if !ok {
		return nil, errors.Errorf("unknown sort rule key - %s", raw)
	}

	return &SortRule{raw, reverse}, nil
}

func (sr *SortRule) equal(t1, t2 *Todo) bool {
	rule, ok := sortRules[sr.Key]
	if !ok {
		return false
	}

	if rule.equal == nil {
		return false
	}

	return rule.equal(t1, t2)
}

func (sr *SortRule) less(t1, t2 *Todo) bool {
	rule, ok := sortRules[sr.Key]
	if !ok {
		return false
	}

	if rule.less == nil {
		return false
	}

	return rule.less(t1, t2)
}

// NewSorter creates a new sorter function from a list of sort rules.
// Sorter functions are indended to be used as argument for the
// sort.SliceStable function.
func NewSorter(
	rules []*SortRule,
) func(s []*Todo) ([]*Todo, func(i, j int) bool) {
	return func(s []*Todo) ([]*Todo, func(i, j int) bool) {
		return s, func(i, j int) bool {
			t1, t2 := s[i], s[j]

			for _, rule := range rules {
				if rule.equal(t1, t2) {
					continue
				}

				if rule.Reverse {
					return rule.less(t2, t1)
				}

				return rule.less(t1, t2)
			}

			return false
		}
	}
}
