package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"rat/config"
	"rat/handler/httputil"
	"rat/logr"
)

var log = logging.MustGetLogger("authhttp")

type handler struct {
	conf *config.AuthConfig
}

// RegisterRoutes registers view routes on given router.
func RegisterRoutes(
	router *mux.Router, log *logr.LogR, conf *config.AuthConfig,
) {
	h := &handler{
		conf: conf,
	}

	authRouter := router.PathPrefix("/auth").
		Subrouter().
		StrictSlash(true)

	authRouter.HandleFunc("/", httputil.Wrap(log, h.auth)).
		Methods(http.MethodGet)
}

func (h *handler) auth(w http.ResponseWriter, r *http.Request) error {
	username, password, ok := r.BasicAuth()
	if !ok {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"failed to parse Authorization header as basic auth",
		)

		return errors.New("failed to parse Authorization header as basic auth")
	}

	if h.conf.Username != username {
		httputil.WriteError(
			w,
			http.StatusUnauthorized,
			"failed to authenticate user",
		)

		return errors.New("failed to authenticate user, user name mismatch")
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(h.conf.PasswordHash), []byte(password),
	)
	if err != nil {
		httputil.WriteError(
			w,
			http.StatusUnauthorized,
			"failed to authenticate user",
		)

		return errors.New("failed to authenticate user, password mismatch")
	}

	token, err := jwt.NewWithClaims(
		jwt.SigningMethodHS512,
		jwt.MapClaims{
			"expires": time.Now().Add(h.conf.TokenExpiration).Unix(),
			"role":    "owner",
		},
	).SignedString([]byte(h.conf.Secret))
	if err != nil {
		httputil.WriteError(
			w, http.StatusInternalServerError, "failed to sign token",
		)

		return errors.Wrap(err, "failed to sign token")
	}

	err = httputil.WriteResponse(
		w,
		http.StatusOK,
		struct {
			Token string `json:"token"`
		}{
			Token: token,
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}

func AuthMW() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerParts := strings.Fields(r.Header.Get("Authorization"))
			if len(headerParts) != 2 ||
				strings.ToLower(headerParts[0]) != "bearer" {
				httputil.WriteError(
					w, http.StatusBadRequest, "invalid Authorization header",
				)

				return
			}

			token, err := jwt.Parse(
				headerParts[1],
				func(token *jwt.Token) (interface{}, error) {
					return nil, nil
					// return secret
				},
				jwt.WithValidMethods([]string{"HS512"}),
			)
		})
	}
}
