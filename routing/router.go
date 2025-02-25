package routing

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
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
	version int
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
	// POC-ONLY
	preDelay := req.Header.Get("x-router-pre-delay")
	postDelay := req.Header.Get("x-router-post-delay")
	slog.Debug("handling router request",
		"path", req.URL.Path,
	)
	if preDelay != "" {
		delayMS, err := strconv.Atoi(preDelay)
		if err != nil {
			http.Error(w, "invalid pre-delay", http.StatusBadRequest)
		}
		time.Sleep(time.Duration(delayMS) * time.Millisecond)
	}
	// /POC-ONLY
	w.Header().Set("X-Router-Version", strconv.Itoa(r.version))
	r.handler.ServeHTTP(w, req)

	// POC-ONLY
	if postDelay != "" {
		delayMS, err := strconv.Atoi(postDelay)
		if err != nil {
			http.Error(w, "invalid post-delay", http.StatusBadRequest)
		}
		time.Sleep(time.Duration(delayMS) * time.Millisecond)
	}
	// /POC-ONLY
}

func (r *Router) Reload() error {
	// TODO some sort of code to ensure only one reload happens at a time

	slog.Info("starting reload of router")
	if r.Storage == nil {
		return errors.New("storage not initialized")
	}
	redirects := r.Storage.GetRedirects()
	slog.Debug("redirects got from storage", "redirects", redirects)
	routes := r.Storage.GetRoutes()
	slog.Debug("routes got from storage", "routes", routes)

	mux := r.newHandler(redirects, routes)

	// TODO it might be worth wrapping these with a mutex to avoid issues with any in-flight requests.
	r.version++
	r.handler = mux

	slog.Info("finished reloading router", "version", strconv.Itoa(r.version))
	return nil
}

func (r *Router) newHandler(redirects []Redirect, routes []Route) http.Handler {
	mux := http.NewServeMux()
	for _, redirect := range redirects {
		slog.Debug("adding redirect", "path", redirect.Path, "redirect", redirect.Redirect, "status", redirect.Status)
		redirectHandler := NewRedirectHandler(redirect.Redirect, redirect.Status)
		mux.Handle(redirect.Path, redirectHandler) // TODO recover panic on bad path
	}
	for _, route := range routes {
		slog.Debug("adding route", "path", route.Path, "host", route.Host)
		proxyHandler, err := NewProxyHandler(route.Host)
		if err != nil {
			panic(err) // TODO handle error properly
		}
		mux.Handle(route.Path, proxyHandler) // TODO recover panic on bad path
	}
	return mux
}
