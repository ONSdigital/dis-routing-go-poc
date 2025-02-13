package upstream

import (
	"log/slog"
	"net/http"
)

func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HandleAny)
	return mux
}

func HandleAny(w http.ResponseWriter, req *http.Request) {
	slog.Info("handling upstream request",
		"path", req.URL.Path,
	)
	w.WriteHeader(http.StatusOK)
}
