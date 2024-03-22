package httputil

import "time"

// DefaultTimeouts is the default timeout values for the api server.
//
//nolint:gochecknoglobals
var DefaultTimeouts = &ServerTimeouts{
	Read:  15 * time.Second,
	Write: 15 * time.Second,
	Idle:  15 * time.Second,
}

// ServerTimeouts defines timeout values for the http server.
type ServerTimeouts struct {
	Read  time.Duration `yaml:"read"`
	Write time.Duration `yaml:"write"`
	Idle  time.Duration `yaml:"idle"`
}

// FillDefaults fills the default values for the timeouts.
func (t *ServerTimeouts) FillDefaults() *ServerTimeouts {
	if t == nil {
		return DefaultTimeouts
	}

	fill := *t

	if fill.Read == 0 {
		fill.Read = DefaultTimeouts.Read
	}

	if fill.Write == 0 {
		fill.Write = DefaultTimeouts.Write
	}

	if fill.Idle == 0 {
		fill.Idle = DefaultTimeouts.Idle
	}

	return &fill
}
