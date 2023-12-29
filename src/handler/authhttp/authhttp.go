package auth

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"rat/graph/services"
	"rat/handler/httputil"
	"rat/logr"
)

type handler struct {
	log *logr.LogR
	gs  *services.GraphServices
}

// RegisterRoutes registers auth routes on given router.
func RegisterRoutes(
	router *mux.Router, log *logr.LogR, gs *services.GraphServices,
) (mux.MiddlewareFunc, error) {
	log = log.Prefix("auth")

	h := &handler{
		log: log,
		gs:  gs,
	}

	authRouter := router.PathPrefix("/auth").Subrouter()

	authRouter.HandleFunc(
		"",
		httputil.Wrap(
			httputil.WrapOptions(
				h.auth,
				[]string{http.MethodPost},
				[]string{"Content-Type"},
			),
			log,
			"auth",
		),
	).Methods(http.MethodPost, http.MethodOptions)

	return h.authMW, nil
}

func (h *handler) auth(w http.ResponseWriter, r *http.Request) error {
	body, err := httputil.Body[struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}](w, r)
	if err != nil {
		return httputil.Error(
			http.StatusBadRequest, errors.Wrap(err, "failed to parse body"),
		)
	}

	signed, _, err := h.gs.Auth.Generate(body.Username, body.Password)
	if err != nil {
		return httputil.Error(
			http.StatusBadRequest, errors.Wrap(err, "failed to generate token"),
		)
	}

	err = httputil.WriteResponse(
		w,
		http.StatusOK,
		struct {
			Token string `json:"token"`
		}{
			Token: signed,
		},
	)
	if err != nil {
		return httputil.Error(
			http.StatusBadRequest,
			errors.Wrap(err, "failed to write response"),
		)
	}

	return nil
}

func (h *handler) authMW(next http.Handler) http.Handler {
	return http.HandlerFunc(httputil.Wrap(
		func(w http.ResponseWriter, r *http.Request) error {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)

				return nil
			}

			signed, err := getToken(r)
			if err != nil {
				return httputil.Error(
					http.StatusBadRequest,
					errors.Wrap(err, "failed to get token"),
				)
			}

			_, err = h.gs.Auth.Validate(signed)
			if err != nil {
				return httputil.Error(
					http.StatusUnauthorized,
					errors.Wrap(err, "failed to validate token"),
				)
			}

			next.ServeHTTP(w, r)

			return nil
		},
		h.log,
		"auth-middleware",
	))
}

func getToken(r *http.Request) (string, error) {
	headerParts := strings.Fields(r.Header.Get("Authorization"))

	if len(headerParts) != 2 {
		return "", errors.New(
			`invalid Authorization header, expected 2 parts - "Bearer <token>"`,
		)
	}

	if !strings.EqualFold("Bearer", headerParts[0]) {
		return "", errors.Errorf(
			`invalid Authorization token kind - %q, expected "Bearer"`,
			headerParts[0],
		)
	}

	return headerParts[1], nil
}
