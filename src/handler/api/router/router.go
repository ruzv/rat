package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"rat/graph"
	"rat/logr"
)

// GraphHandlerFunc defines a handler function signature for HTTP request
// that request handler functons that interact with the graph.
type GraphHandlerFunc func(
	p graph.Provider, w http.ResponseWriter, r *http.Request,
) error

type Router struct {
	log      *logr.LogR
	provider graph.Provider
	muxRouter   *mux.Router
}

type Handler struct {
    router *Router
    method string

}

func NewRouter(log *logr.LogR, provider graph.Provider) (*Router, error) {
	log = log.Prefix("router")

	// router := mux.NewRouter()

	return &Router{
		log:      log,
		provider: provider,
		muxRouter:   mux.NewRouter(),
	}, nil
}

func (r *Router) Path(path string) *Router {
    newMuxRouter := r.muxRouter.PathPrefix(path).Subrouter()
    
    return &Router{

        muxRouter: newMuxRouter,
    }
}

func (r *Router) Method(method string) *Handler {
    return &Handler{
        router: r,
        method: method,
    }
}

func (h *Handler) GraphHandler(
    handler GraphHandlerFunc,
) {
    r.router.HandleFunc(path, r.graphAccessWrapper(r.provider, r.log, handler))
}


func (r *Router) graphAccessWrapper(
	base graph.Provider,
	log *logr.LogR,
	handler GraphHandlerFunc,
) httputil.RatHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		token, ok := r.Context().Value(auth.AuthTokenCtxKey).(*auth.Token)
		if !ok || token == nil {
			return httputil.Error(
				http.StatusInternalServerError,
				errors.New("failed to get auth token from context"),
			)
		}

		b, err := json.MarshalIndent(token, "", "  ")
		if err != nil {
			return httputil.Error(
				http.StatusInternalServerError,
				errors.Wrap(err, "failed to marshal token"),
			)
		}

		log.Debugf("token:\n%s", string(b))

		err = handler(access.NewProvider(base, log, token.Scopes), w, r)
		if err != nil {
			return err
		}

		return nil
	}
}


