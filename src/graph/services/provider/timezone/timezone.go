package timezone

import (
	"time"

	"github.com/pkg/errors"
	"rat/graph"
)

// Provider wraps read/write provider with a timezone getter method,
// implementing the complete graph Provider.
type Provider struct {
	graph.ReadWriteProvider
	loc *time.Location
}

// New initalses a new time zone provider.
func New(
	p graph.ReadWriteProvider,
	timeZone string,
) (*Provider, error) {
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to load location for timezone %q",
			timeZone,
		)
	}

	return &Provider{
		ReadWriteProvider: p,
		loc:               loc,
	}, nil
}

// TimeZone returns the graphs time zone.
func (p *Provider) TimeZone() *time.Location {
	loc := *p.loc

	return &loc
}
