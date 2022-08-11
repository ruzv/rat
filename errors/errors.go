package errors

import (
	"fmt"
	"runtime"
)

var _ error = (*ErrWithTrace)(nil)

// type Err struct {
// 	errors []error
// }

type ErrWithTrace struct {
	line    int
	file    string
	message string
	err     error
}

func (e *ErrWithTrace) Error() string {
	var errMessage string
	if e.err != nil {
		errMessage = e.err.Error()
	}

	return fmt.Sprintf("%s:%d %s\n%s", e.file, e.line, e.message, errMessage)
}

func New(message string) error {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}

	return &ErrWithTrace{
		line:    line,
		file:    file,
		message: message,
	}
}

func Wrap(err error, message string) error {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}

	return &ErrWithTrace{
		line:    line,
		file:    file,
		message: message,
		err:     err,
	}
}

func Wrapf(err error, format string, args ...interface{}) error {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}

	return &ErrWithTrace{
		line:    line,
		file:    file,
		message: fmt.Sprintf(format, args...),
		err:     err,
	}
}
