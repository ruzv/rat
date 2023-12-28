package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"rat/graph/services/auth"
	"rat/graph/util"
	"rat/handler/httputil"
	"rat/logr"
)

type handler struct {
	log         *logr.LogR
	users       map[string]*auth.User
	tokenConfig *auth.TokenConfig
}

// RegisterRoutes registers auth routes on given router.
func RegisterRoutes(
	router *mux.Router,
	log *logr.LogR,
	users []*auth.User,
	tokenConfig *auth.TokenConfig,
) (mux.MiddlewareFunc, error) {
	log = log.Prefix("auth")

	h := &handler{
		log: log,
		users: util.ObjectMap(
			users,
			func(u *auth.User) string { return u.Username },
		),
		tokenConfig: tokenConfig,
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

	user, ok := h.users[body.Username]
	if !ok {
		return httputil.Error(
			http.StatusBadRequest,
			errors.New("failed to authenticate user"),
		)
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(body.Password),
	)
	if err != nil {
		h.log.Warnf(
			"failed to authenticate user - %q: %s", user.Username, err.Error(),
		)

		return httputil.Error(
			http.StatusBadRequest,
			errors.New("failed to authenticate user"),
		)
	}

	token, err := jwt.NewWithClaims(
		jwt.SigningMethodHS512,
		auth.Token{
			Username: user.Username,
			Expires:  time.Now().Add(h.tokenConfig.Expiration).Unix(),
			Role:     user.Role,
		}.ToMapClaims(),
	).SignedString([]byte(h.tokenConfig.Secret))
	if err != nil {
		return httputil.Error(
			http.StatusBadRequest,
			errors.Wrap(err, "failed to sign token"),
		)
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
		return httputil.Error(
			http.StatusBadRequest,
			errors.Wrap(err, "failed to write response"),
		)
	}

	return nil
}

func (h *handler) authMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)

			return
		}

		headerParts := strings.Fields(r.Header.Get("Authorization"))

		if len(headerParts) != 2 ||
			strings.EqualFold("Bearer", headerParts[0]) {
			h.log.Debugf(
				"malformed authorization header: %q",
				r.Header.Get("Authorization"),
			)

			httputil.WriteError(
				w, http.StatusBadRequest, "invalid Authorization header",
			)

			return
		}

		token, err := jwt.Parse(
			headerParts[1],
			func(token *jwt.Token) (any, error) {
				return []byte(h.tokenConfig.Secret), nil
			},
			jwt.WithValidMethods([]string{"HS512"}),
		)
		if err != nil {
			h.log.Debugf("failed to parse token: %s", err.Error())

			httputil.WriteError(
				w, http.StatusBadRequest, "invalid authorization token",
			)

			return
		}

		if !token.Valid {
			h.log.Debugf("token not valid")

			httputil.WriteError(
				w, http.StatusBadRequest, "invalid authorization token",
			)

			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			h.log.Debugf("failed to parse claims")

			httputil.WriteError(
				w, http.StatusBadRequest, "invalid authorization token",
			)

			return
		}

		tc, err := auth.FromMapClaims(claims)
		if err != nil {
			h.log.Debugf("failed to parse token claims")

			httputil.WriteError(
				w, http.StatusBadRequest, "invalid authorization token",
			)

			return
		}

		if tc.Expired() {
			h.log.Debugf(
				"token expired, expires - %d, now - %d",
				tc.Expires,
				time.Now().Unix(),
			)

			httputil.WriteError(
				w, http.StatusUnauthorized, "token expired",
			)

			return
		}

		next.ServeHTTP(w, r)
	})
}
