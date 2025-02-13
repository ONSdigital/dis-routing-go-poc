package storage

import (
	"log/slog"
	"net/http"
)

var _ http.Handler = &Store{}

func (s Store) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.handler.ServeHTTP(writer, request)
}

func adminHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleAny)
	return mux
}

func handleAny(w http.ResponseWriter, req *http.Request) {
	slog.Info("handling admin request",
		"path", req.URL.Path,
	)
}
