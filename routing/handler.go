package routing

import (
	"log/slog"
	"net/http"
)

type Handler struct {
}

func (h Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	slog.Info("handling router request",
		"path", req.URL.Path,
	)
}
