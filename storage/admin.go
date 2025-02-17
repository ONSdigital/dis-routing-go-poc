package storage

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/ONSdigital/dis-routing-go-poc/routing"
)

func (s *Store) AdminHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /routes", s.addRoute)
	mux.HandleFunc("POST /redirects", s.addRedirect)
	return mux
}

type addRouteModel struct {
	Path string `json:"path"`
	Host string `json:"host"`
}

func (s *Store) addRoute(w http.ResponseWriter, req *http.Request) {
	slog.Info("received add route request")
	body, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("error reading body")
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	var payload addRouteModel
	if err := json.Unmarshal(body, &payload); err != nil {
		slog.Error("error unmarshalling body")
		http.Error(w, "invalid json in request", http.StatusBadRequest)
		return
	}

	// TODO VALIDATE INPUTS!!!!

	slog.Info("adding route", "path", payload.Path, "host", payload.Host)
	route := routing.Route{
		Path: payload.Path,
		Host: payload.Host,
	}

	s.routes[payload.Path] = route

	if s.reloadFunc != nil {
		if err := s.reloadFunc(); err != nil {
			slog.Error("error reloading router")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

type addRedirectModel struct {
	Path     string `json:"path"`
	Redirect string `json:"redirect"`
	Type     string `json:"type"`
}

func (s *Store) addRedirect(w http.ResponseWriter, req *http.Request) {
	slog.Info("received add redirect request")
	body, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("error reading body")
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	var payload addRedirectModel
	if err := json.Unmarshal(body, &payload); err != nil {
		slog.Error("error unmarshalling body")
		http.Error(w, "invalid json in request", http.StatusBadRequest)
		return
	}

	// TODO VALIDATE INPUTS!!!!

	var status int
	switch payload.Type {
	case "perm":
		status = http.StatusPermanentRedirect
	case "temp":
		status = http.StatusTemporaryRedirect
	default:
		status = http.StatusTemporaryRedirect
	}

	slog.Info("adding redirect", "path", payload.Path, "redirect", payload.Redirect, "type", payload.Type, "status", status)
	redirect := routing.Redirect{
		Path:     payload.Path,
		Redirect: payload.Redirect,
		Status:   status,
	}

	s.redirects[payload.Path] = redirect

	if s.reloadFunc != nil {
		if err := s.reloadFunc(); err != nil {
			slog.Error("error reloading router")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}
