package token

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"rat/graph/render/jsonast"
)

// ErrUnknownTimeTokenOperation unknown time token opp.
var ErrUnknownTimeTokenOperation = errors.New("unknown opperation")

func (t *Token) renderTime(
	part *jsonast.AstPart,
) error {
	opperation, ok := t.Args["opp"]
	if !ok {
		return errors.Wrap(
			ErrMissingArgument,
			"time 'opp' argument must have value one of 'now', `since`",
		)
	}

	switch opperation {
	case "now":
		format, ok := t.Args["format"]
		if !ok {
			format = "02.01.2006 15:04"
		}

		part.AddLeaf(
			&jsonast.AstPart{
				Type: "text",
				Attributes: jsonast.AstAttributes{
					"text": time.Now().Format(format),
				},
			},
		)

		return nil
	case "since":
		dateArg, ok := t.Args["date"]
		if !ok {
			return errors.Wrap(ErrMissingArgument, "missing 'date'")
		}

		date, err := time.Parse("02.01.2006", dateArg)
		if err != nil {
			return errors.Wrap(err, "failed to parse date form arg")
		}

		part.AddLeaf(
			&jsonast.AstPart{
				Type: "text",
				Attributes: jsonast.AstAttributes{
					"text": fmt.Sprintf("%.2f", time.Since(date).Hours()/24),
				},
			},
		)

		return nil
	default:
		return errors.Wrapf(
			ErrUnknownTimeTokenOperation,
			"operation %q",
			opperation,
		)
	}
}
