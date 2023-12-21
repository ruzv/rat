package httputil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"rat/logr"
)

var (
	_ http.ResponseWriter = (*BufferResponseWriter)(nil)

	_ error = (*httpError)(nil) //nolint:errcheck // if err != lol
)

type (
	// MuxHandlerFunc is a handler function for mux.
	MuxHandlerFunc func(http.ResponseWriter, *http.Request)
	// RatHandlerFunc is a handler function.
	RatHandlerFunc func(http.ResponseWriter, *http.Request) error
)

// BufferResponseWriter is a wrapper around http.ResponseWriter that proxies
// the data ans stores the status code.
type BufferResponseWriter struct {
	Code int
	w    http.ResponseWriter
}

type httpError struct {
	statusCode int
	err        error
}

// NewBufferResponseWriter creates a new buffered response writer.
func NewBufferResponseWriter(w http.ResponseWriter) *BufferResponseWriter {
	return &BufferResponseWriter{w: w}
}

// Header returns the header.
func (w *BufferResponseWriter) Header() http.Header {
	return w.w.Header()
}

// Write writes the data to the buffer.
func (w *BufferResponseWriter) Write(b []byte) (int, error) {
	n, err := w.w.Write(b)
	if err != nil {
		return 0, errors.Wrap(err, "failed to write")
	}

	return n, nil
}

// WriteHeader sets the status code.
func (w *BufferResponseWriter) WriteHeader(code int) {
	w.Code = code
	w.w.WriteHeader(code)
}

// Wrap wraps a RatHandlerFunc to be used with mux.
func Wrap(log *logr.LogR, f RatHandlerFunc) MuxHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			log.Errorf("handler error: %v", err)

			httpErr := &httpError{}
			if errors.As(err, &httpErr) {
				WriteError(w, httpErr.statusCode, httpErr.err.Error())
			}
		}
	}
}

// WriteResponse writes a response to the response writer.
func WriteResponse(w http.ResponseWriter, code int, body any) error {
	b := &bytes.Buffer{}

	err := json.NewEncoder(b).Encode(body)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to encode body")

		return errors.Wrap(err, "failed to encode body")
	}

	w.WriteHeader(code)

	_, err = w.Write(b.Bytes())
	if err != nil {
		return errors.Wrap(err, "failed to write body")
	}

	return nil
}

// WriteError writes an error to the response.
func WriteError(w http.ResponseWriter, code int, format string, args ...any) {
	WriteResponse( //nolint:errcheck,gosec // avoid double error handling.
		w,
		code,
		struct {
			Code  int    `json:"code"`
			Error string `json:"error"`
		}{
			Code:  code,
			Error: fmt.Sprintf(format, args...),
		},
	)
}

// Body reads the requests body as a specified struct.
func Body[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	defer r.Body.Close() //nolint:errcheck // ignore.

	var body, empty T

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "failed to decode body")

		return empty, errors.Wrap(err, "failed to decode body")
	}

	return body, nil
}

// Error returns the error message of HTTP error.
func (e *httpError) Error() string {
	return fmt.Sprintf(
		"HTTP error %d %s: %s",
		e.statusCode,
		http.StatusText(e.statusCode),
		e.err.Error(),
	)
}

// Error creates a new HTTP error, that can be handled by Wrap function, writing
// the error message and status code to the response.
func Error(statusCode int, err error) error {
	return &httpError{
		statusCode: statusCode,
		err:        err,
	}
}
