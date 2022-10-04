package handler

import (
	"encoding/json"
	"net/http"

	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var log = logging.MustGetLogger("handler-utils")

type (
	MuxHandlerFunc func(http.ResponseWriter, *http.Request)
	RatHandlerFunc func(http.ResponseWriter, *http.Request) error
)

func Wrap(f RatHandlerFunc) MuxHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			log.Errorf("handler error: %v", err)
		}
	}
}

func WriteResponse(w http.ResponseWriter, code int, body interface{}) error {
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		return errors.Wrap(err, "failed to encode body")
	}

	return nil
}

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
