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
)

// Hint represents a todo hint. That can have a type and a value.
type Hint struct {
	Type  HintType
	Value interface{}
}

var errUnknownHint = errors.New("unknown hint type")

var hintsTypeProcessors = map[HintType]*struct {
	parse  func(string) (*Hint, error)
	format func(interface{}) string
}{
	Due: {
		parse: func(s string) (*Hint, error) {
			due, err := parseTime(s)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse due date")
			}

			return &Hint{Due, due}, nil
		},
		format: func(v interface{}) string {
			t, ok := v.(time.Time)
			if !ok {
				return fmt.Sprintf("%v", v)
			}

			return t.Format("02.01.2006")
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
	},
	Src: {},
}

func parseHint(line string) (*Hint, error) {
	parts := strings.Split(line, "=")

	if len(parts) != 2 {
		return nil, errors.Errorf("invalid hint - %q", line)
	}

	hType := strings.TrimSpace(parts[0])
	hValue := strings.TrimSpace(parts[1])

	p, ok := hintsTypeProcessors[HintType(hType)]
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

// String returns a string representation of the hint.
func (h *Hint) String() string {
	if h.Type == None {
		return ""
	}

	p, ok := hintsTypeProcessors[h.Type]
	if !ok {
		return fmt.Sprintf("%v", h.Value)
	}

	if p.format == nil {
		return fmt.Sprintf("%v", h.Value)
	}

	return p.format(h.Value)
}

func (h *Hint) markdown() string {
	return fmt.Sprintf("%s = %s", h.Type, h.String())
}
