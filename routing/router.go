package routing

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type StorageReadWriter interface {
	GetRedirects() []Redirect
}

type Redirect struct {
	Path     string
	Redirect string
	Status   int
}

type Router struct {
	Storage StorageReadWriter
	handler http.Handler
}

func NewRouter(storage StorageReadWriter) (*Router, error) {
	router := &Router{
		Storage: storage,
	}
	if err := router.Reload(); err != nil {
		return nil, fmt.Errorf("error initially reloading router: %w", err)
	}
	return router, nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

func (r *Router) Reload() error {
	slog.Info("starting reload of router")
	if r.Storage == nil {
		return errors.New("storage not initialized")
	}
	redirects := r.Storage.GetRedirects()
	slog.Debug("redirects got from storage", "redirects", redirects)

	mux := r.newHandler(redirects)
	r.handler = mux
	slog.Info("finished reloading router")
	return nil
}

func (r *Router) newHandler(redirects []Redirect) http.Handler {
	mux := http.NewServeMux()
	for _, redirect := range redirects {
		slog.Debug("adding redirect", "path", redirect.Path, "redirect", redirect.Redirect, "status", redirect.Status)
		mux.HandleFunc(redirect.Path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Location", redirect.Redirect)
			w.WriteHeader(redirect.Status)
		})
	}
	return mux
}
