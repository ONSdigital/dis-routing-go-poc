package storage

import (
	"log/slog"
	"net/http"
)

func AdminHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HandleAny)
	return mux
}

func HandleAny(w http.ResponseWriter, req *http.Request) {
	slog.Info("handling admin request",
		"path", req.URL.Path,
	)
}
