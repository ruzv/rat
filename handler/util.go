package handler

import (
	"encoding/json"
	"net/http"

	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var log = logging.MustGetLogger("handler-utils")

type (
	// MuxHandlerFunc is a handler function for mux.
	MuxHandlerFunc func(http.ResponseWriter, *http.Request)
	// RatHandlerFunc is a handler function.
	RatHandlerFunc func(http.ResponseWriter, *http.Request) error
)

// Wrap wraps a RatHandlerFunc to be used with mux.
func Wrap(f RatHandlerFunc) MuxHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			log.Errorf("handler error: %v", err)
		}
	}
}

// WriteResponse writes a response to the response writer.
func WriteResponse(w http.ResponseWriter, code int, body any) error {
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		return errors.Wrap(err, "failed to encode body")
	}

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
func Body[T any](w http.ResponseWriter, r *http.Request) (T, error) { //nolint:ireturn,lll
	defer r.Body.Close()

	var body, empty T

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "failed to decode body")

		return empty, errors.Wrap(err, "failed to decode body")
	}

	return body, nil
}
