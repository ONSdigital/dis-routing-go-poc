package routing

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type StorageReadWriter interface {
	GetRedirects() []Redirect
	GetRoutes() []Route
	RegisterReloadFunc(func() error)
}

type Redirect struct {
	Path     string
	Redirect string
	Status   int
}

type Route struct {
	Path string
	Host string
}

type Router struct {
	Storage StorageReadWriter
	handler http.Handler
}

func NewRouter(storage StorageReadWriter) (*Router, error) {
	router := &Router{
		Storage: storage,
	}
	storage.RegisterReloadFunc(router.Reload)
	if err := router.Reload(); err != nil {
		return nil, fmt.Errorf("error initially reloading router: %w", err)
	}
	return router, nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	slog.Info("handling router request",
		"path", req.URL.Path,
	)
	r.handler.ServeHTTP(w, req)
}

func (r *Router) Reload() error {
	slog.Info("starting reload of router")
	if r.Storage == nil {
		return errors.New("storage not initialized")
	}
	redirects := r.Storage.GetRedirects()
	slog.Debug("redirects got from storage", "redirects", redirects)
	routes := r.Storage.GetRoutes()
	slog.Debug("routes got from storage", "routes", routes)

	mux := r.newHandler(redirects, routes)
	r.handler = mux
	slog.Info("finished reloading router")
	return nil
}

func (r *Router) newHandler(redirects []Redirect, routes []Route) http.Handler {
	mux := http.NewServeMux()
	for _, redirect := range redirects {
		slog.Debug("adding redirect", "path", redirect.Path, "redirect", redirect.Redirect, "status", redirect.Status)
		redirectHandler := NewRedirectHandler(redirect.Path, redirect.Status)
		mux.Handle(redirect.Path, redirectHandler)
	}
	for _, route := range routes {
		slog.Debug("adding route", "path", route.Path, "host", route.Host)
		proxyHandler, err := NewProxyHandler(route.Host)
		if err != nil {
			panic(err)
		}
		mux.Handle(route.Path, proxyHandler)
	}
	return mux
}
