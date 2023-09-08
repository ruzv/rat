package httputil

import (
	"bytes"
	"encoding/json"
	"net/http"

	"rat/logr"

	"github.com/pkg/errors"
)

var _ http.ResponseWriter = (*BufferResponseWriter)(nil)

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

// NewBufferResponseWriter creates a new buffered response writer.
func NewBufferResponseWriter(w http.ResponseWriter) *BufferResponseWriter {
	return &BufferResponseWriter{w: w}
}

// Header .
func (w *BufferResponseWriter) Header() http.Header {
	return w.w.Header()
}

// Write .
func (w *BufferResponseWriter) Write(b []byte) (int, error) {
	n, err := w.w.Write(b)
	if err != nil {
		return 0, errors.Wrap(err, "failed to write")
	}

	return n, nil
}

// WriteHeader .
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
	w.Write(b.Bytes()) //nolint:errcheck

	return nil
}

// WriteError writes an error to the response.
func WriteError(w http.ResponseWriter, code int, message string) {
	WriteResponse( //nolint:errcheck
		w,
		code,
		struct {
			Code  int    `json:"code"`
			Error string `json:"error"`
		}{
			Code:  code,
			Error: message,
		},
	)
}

// Body reads the requests body as a specified struct.
func Body[T any](
	w http.ResponseWriter,
	r *http.Request,
) (T, error) { //nolint:ireturn
	defer r.Body.Close()

	var body, empty T

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "failed to decode body")

		return empty, errors.Wrap(err, "failed to decode body")
	}

	return body, nil
}
