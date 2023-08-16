package todo

import (
	"strings"

	"rat/graph/util"

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

// FilterHas performs fltration of a list of todos by a list of
// filter rules.
func FilterHas(todos []*Todo, rules []*FilterRule) []*Todo {
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

// Operator represents a filtering by value operator.
type Operator string

const (
	// Equal represents an equal operator.
	Equal Operator = "="
	// Less represents a less operator.
	Less Operator = "<"
	// Greater represents a greater operator.
	Greater Operator = ">"
)

// FilterValueRule represents a single filtering rule of multiple todos that
// will check if a certain hints value is equal, less or greater than a given
// value.
type FilterValueRule struct {
	Type     HintType
	Operator Operator
	Value    any
}

// ParseFilterValueRule parses a filter value rule from a raw string.
func ParseFilterValueRule(raw string) (*FilterValueRule, error) {
	var operator string

	for _, op := range []string{"=", "<", ">"} {
		if strings.Contains(raw, op) {
			operator = op

			break
		}
	}

	if operator == "" {
		return nil, errors.New("no operator found")
	}

	parts := strings.Split(raw, operator)

	p, ok := hintTypeProcessors[HintType(parts[0])]
	if !ok {
		return nil, errors.Errorf("unknown hint type - %s", parts[0])
	}

	if p.parseFilterValue != nil {
		v, err := p.parseFilterValue(parts[1])
		if err != nil {
			return nil,
				errors.Wrapf(err, "failed to parse filter value %q", parts[1])
		}

		return &FilterValueRule{HintType(parts[0]), Operator(operator), v}, nil
	}

	if p.parse != nil {
		v, err := p.parse(parts[1])
		if err != nil {
			return nil,
				errors.Wrapf(err, "failed to parse filter value %q", parts[1])
		}

		return &FilterValueRule{HintType(parts[0]), Operator(operator), v}, nil
	}

	return &FilterValueRule{HintType(parts[0]), Operator(operator), parts[1]},
		nil
}

// FilterValue performs fltration of a list of todos by a list of filter value
// rules.
func FilterValue(todos []*Todo, rules []*FilterValueRule) []*Todo {
	for _, r := range rules {
		todos = util.Filter(
			todos,
			func(t *Todo) bool {
				p, ok := hintTypeProcessors[r.Type]
				if !ok {
					return false
				}

				var operator func(any, any) bool

				switch r.Operator {
				case Equal:
					operator = p.equal
				case Less:
					operator = p.less
				case Greater:
					operator = p.greater
				default:
					return false
				}

				if operator == nil {
					return true
				}

				h := t.GetHint(r.Type)
				if h == nil {
					return false
				}

				return operator(h.Value, r.Value)
			},
		)
	}

	return todos
}
