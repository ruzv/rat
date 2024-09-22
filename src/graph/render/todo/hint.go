package todo

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	// None is a hint type that does not provide any hints.
	None HintType = "none"
	// Due is a hint type that provides a due date.
	Due HintType = "due"
	// Size is a hint type that provides a size.
	Size HintType = "size"
	// Src is a hint type that shows the source node path of todo.
	Src HintType = "src"
	// Priority is a hint type that specifies the priority of a todo.
	Priority HintType = "priority"
	// Tags is a hint type that specifies a list of tags for a todo.
	Tags HintType = "tags"
)

var errUnknownHint = errors.New("unknown hint type")

// HintType describes todo hint types.
type HintType string

// Hint represents a todo hint. That can have a type and a value.
type Hint struct {
	Type  HintType `json:"type"`
	Value any      `json:"value"`
}

type comparableValue interface {
	int | time.Duration
}

//nolint:gochecknoglobals,decorder // TODO: fix.
var hintTypeProcessors = map[HintType]*struct {
	parse            func(string) (any, error)
	parseFilterValue func(string) (any, error)
	formatMD         func(any) string
	formatHTML       func(any) string
	//           self, other
	equal   func(any, any) bool
	less    func(any, any) bool
	greater func(any, any) bool
}{
	Due: {
		parse: func(s string) (any, error) {
			due, err := parseTime(s)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse due date")
			}

			return due, nil
		},
		parseFilterValue: func(s string) (any, error) {
			v, err := strconv.Atoi(s)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse due filter value")
			}

			return v, nil
		},
		formatMD: func(v any) string {
			t, ok := v.(time.Time)
			if !ok {
				return fmt.Sprintf("%v", v)
			}

			return t.Format("02.01.2006")
		},
		formatHTML: func(v any) string {
			t, ok := v.(time.Time)
			if !ok {
				return fmt.Sprintf("%v", v)
			}

			return fmt.Sprintf(
				"%s in %.2f days",
				t.Format("02.01.2006"),
				time.Until(t).Hours()/24,
			)
		},
		equal: func(self, other any) bool {
			t, ok := self.(time.Time)
			if !ok {
				return false
			}

			d, ok := other.(int)
			if !ok {
				return false
			}

			return int(time.Until(t).Hours()/24) == d
		},
		less: func(self, other any) bool {
			t, ok := self.(time.Time)
			if !ok {
				return false
			}

			d, ok := other.(int)
			if !ok {
				return false
			}

			return int(time.Until(t).Hours()/24) < d
		},
		greater: func(self, other any) bool {
			t, ok := self.(time.Time)
			if !ok {
				return false
			}

			d, ok := other.(int)
			if !ok {
				return false
			}

			return int(time.Until(t).Hours()/24) > d
		},
	},
	Size: {
		parse: func(s string) (any, error) {
			size, err := time.ParseDuration(s)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse size")
			}

			return size, nil
		},
		formatMD: func(v any) string {
			d, ok := v.(time.Duration)
			if !ok {
				return fmt.Sprintf("%v", v)
			}

			f := ""

			h := d / time.Hour
			m := (d % time.Hour) / time.Minute
			s := (d % time.Minute) / time.Second

			if h > 0 {
				f += fmt.Sprintf("%dh", h)
			}

			if m > 0 {
				f += fmt.Sprintf("%dm", m)
			}

			if s > 0 {
				f += fmt.Sprintf("%ds", s)
			}

			return f
		},
		equal:   equal[time.Duration],
		less:    less[time.Duration],
		greater: greater[time.Duration],
	},
	Src: {
		formatHTML: func(v any) string {
			s, ok := v.(string)
			if !ok {
				return fmt.Sprintf("%v", v)
			}

			return s
		},
	},
	Priority: {
		parse: func(s string) (any, error) {
			p, err := strconv.Atoi(s)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse priority")
			}

			return p, nil
		},
		equal:   equal[int],
		less:    less[int],
		greater: greater[int],
	},
	Tags: {
		equal: func(self, other any) bool {
			s, ok := self.(string)
			if !ok {
				return false
			}

			o, ok := other.(string)
			if !ok {
				return false
			}

			for _, tag := range strings.Split(s, ",") {
				if tag == o {
					return true
				}
			}

			return false
		},
	},
}

func parseHint(line string) (*Hint, error) {
	parts := strings.Split(line, "=")

	if len(parts) != 2 {
		return nil, errors.Errorf("invalid hint - %q", line)
	}

	hType := strings.TrimSpace(parts[0])
	hValue := strings.TrimSpace(parts[1])

	p, ok := hintTypeProcessors[HintType(hType)]
	if !ok {
		return nil, errors.Wrapf(
			errUnknownHint,
			"unknown hint type %q, with value %q",
			hType,
			hValue,
		)
	}

	if p.parse == nil {
		return &Hint{HintType(hType), hValue}, nil
	}

	v, err := p.parse(hValue)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse hint value %q", hValue)
	}

	return &Hint{HintType(hType), v}, nil
}

// Value returns the value of the hint as the given type. Returns an empty
// value of type T if hint is nil or hint value is not of type T.
func Value[T any](h *Hint) T { //nolint:ireturn
	var empty T

	if h == nil {
		return empty
	}

	v, ok := h.Value.(T)
	if !ok {
		return empty
	}

	return v
}

func equal[T comparableValue](self, other any) bool {
	s, ok := self.(T)
	if !ok {
		return false
	}

	o, ok := other.(T)
	if !ok {
		return false
	}

	return s == o
}

func less[T comparableValue](self, other any) bool {
	s, ok := self.(T)
	if !ok {
		return false
	}

	o, ok := other.(T)
	if !ok {
		return false
	}

	return s < o
}

func greater[T comparableValue](self, other any) bool {
	s, ok := self.(T)
	if !ok {
		return false
	}

	o, ok := other.(T)
	if !ok {
		return false
	}

	return s > o
}
