package routing

import (
	"log/slog"
	"net/http"
)

type RedirectHandler struct {
	Destination string
	StatusCode  int
}

var _ http.Handler = &RedirectHandler{}

func NewRedirectHandler(destination string, statusCode int) *RedirectHandler {
	return &RedirectHandler{
		Destination: destination,
		StatusCode:  statusCode,
	}
}

func (r *RedirectHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Location", r.Destination)
	w.WriteHeader(r.StatusCode)
	slog.Debug("redirecting request", "from", req.URL.Path, "to", r.Destination)
}
