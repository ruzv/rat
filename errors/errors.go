package errors

import (
	"fmt"
	"runtime"
)

const filenameUnknown = "unknown"

var _ error = (*TraceError)(nil)

// ErrWithTrace wrapper struct for error that also implements error.
// Hods additional trace information.
type TraceError struct {
	line    int
	file    string
	message string
	err     error
}

// Error converts an error to string.
func (e *TraceError) Error() string {
	var errMessage string
	if e.err != nil {
		errMessage = e.err.Error()
	}

	return fmt.Sprintf("%s:%d %s\n%s", e.file, e.line, e.message, errMessage)
}

// New creates a new error with the given message and adds trace information.
func New(message string) error {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = filenameUnknown
		line = 0
	}

	return &TraceError{
		line:    line,
		file:    file,
		message: message,
	}
}

// Wrap wraps an error with trace information.
func Wrap(err error, message string) error {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = filenameUnknown
		line = 0
	}

	return &TraceError{
		line:    line,
		file:    file,
		message: message,
		err:     err,
	}
}

// Wrapf wraps an error with trace information.
func Wrapf(err error, format string, args ...interface{}) error {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = filenameUnknown
		line = 0
	}

	return &TraceError{
		line:    line,
		file:    file,
		message: fmt.Sprintf(format, args...),
		err:     err,
	}
}
