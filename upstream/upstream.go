package upstream

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HandleAny)
	return mux
}

func HandleAny(w http.ResponseWriter, req *http.Request) {
	delay := req.Header.Get("x-upstream-delay")
	slog.Debug("handling upstream request",
		"path", req.URL.Path,
		"delay", delay,
	)
	if delay != "" {
		delayMS, err := strconv.Atoi(delay)
		if err != nil {
			http.Error(w, "invalid delay", http.StatusBadRequest)
		}
		time.Sleep(time.Duration(delayMS) * time.Millisecond)
	}
	w.Header().Set("x-upstream", "true")
	w.WriteHeader(http.StatusOK)
}
