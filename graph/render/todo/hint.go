package todo

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// HintType describes todo hint types.
type HintType string

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

// Hint represents a todo hint. That can have a type and a value.
type Hint struct {
	Type  HintType
	Value any
}

var hintTypeProcessors = map[HintType]*struct {
	parse      func(string) (*Hint, error)
	formatMD   func(any) string
	formatHTML func(any) string
}{
	Due: {
		parse: func(s string) (*Hint, error) {
			due, err := parseTime(s)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse due date")
			}

			return &Hint{Due, due}, nil
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

			// &nbs; - non-breaking space
			return fmt.Sprintf(
				"%s in %.2f days",
				t.Format("02.01.2006"),
				time.Until(t).Hours()/24,
			)
		},
	},
	Size: {
		parse: func(s string) (*Hint, error) {
			size, err := time.ParseDuration(s)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse size")
			}

			return &Hint{Size, size}, nil
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
	},
	Src: {
		formatHTML: func(v any) string {
			s, ok := v.(string)
			if !ok {
				return fmt.Sprintf("%v", v)
			}

			return strings.ReplaceAll(s, "-", "&#8209;")
		},
	},
	Priority: {},
	Tags:     {},
}

var errUnknownHint = errors.New("unknown hint type")

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

	return p.parse(hValue)
}

// HTML returns a string representation of the hints value as it is presented
// in HTML.
func (h *Hint) HTML() string {
	if h.Type == None {
		return ""
	}

	p, ok := hintTypeProcessors[h.Type]
	if !ok {
		return fmt.Sprintf("%v", h.Value)
	}

	if p.formatHTML == nil {
		if p.formatMD == nil {
			return fmt.Sprintf("%v", h.Value)
		}

		return p.formatMD(h.Value)
	}

	return p.formatHTML(h.Value)
}

func (h *Hint) markdown() string {
	if h.Type == None {
		return ""
	}

	p, ok := hintTypeProcessors[h.Type]
	if !ok {
		return fmt.Sprintf("%s = %v", h.Type, h.Value)
	}

	if p.formatMD == nil {
		return fmt.Sprintf("%s = %v", h.Type, h.Value)
	}

	return fmt.Sprintf("%s = %s", h.Type, p.formatMD(h.Value))
}

// Value returns the value of the hint as the given type. Returns an empty
// value of type T if hint is nil or hint value is not of type T.
func Value[T any](h *Hint) T {
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
