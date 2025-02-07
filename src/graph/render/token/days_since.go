package token

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/render/jsonast"
)

func (t *Token) renderTimeSince(
	part *jsonast.AstPart,
	p graph.Provider,
) error {
	dateArg, ok := t.Args["date"]
	if !ok {
		return errors.Wrap(
			ErrMissingArgument, "missing date arg for days since token",
		)
	}

	// all date parsing
	// here
	// in todo token
	// in template fields

	// time.ParseInLocation

	date, err := time.ParseInLocation("02.01.2006", dateArg, p.TimeZone())
	if err != nil {
		return errors.Wrap(err, "failed to parse date form arg")
	}

	// fmt.Println(date.Format("02-01-2006 15:04:05.00000"))
	//
	//    loc, err := time.LoadLocation("Europe/Riga")
	//
	//    time.ParseInLocation

	part.AddLeaf(
		&jsonast.AstPart{
			Type: "text",
			Attributes: jsonast.AstAttributes{
				"text": fmt.Sprintf("%.2f", time.Since(date).Hours()/24),
			},
		},
	)

	return nil
}

func (t *Token) renderTimeNow(
	part *jsonast.AstPart,
	p graph.Provider,
) {
	format, ok := t.Args["format"]
	if !ok {
		format = time.RFC3339
	}

	part.AddLeaf(
		&jsonast.AstPart{
			Type: "text",
			Attributes: jsonast.AstAttributes{
				"text": time.Now().In(p.TimeZone()).Format(format),
			},
		},
	)
}
