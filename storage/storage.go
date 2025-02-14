package storage

import (
	"maps"
	"net/http"
	"slices"

	"github.com/ONSdigital/dis-routing-go-poc/routing"
)

type Store struct {
	reloadFunc func() error
	routes     map[string]routing.Route
	redirects  map[string]routing.Redirect
}

func NewStore() *Store {
	store := &Store{
		redirects: make(map[string]routing.Redirect),
		routes:    make(map[string]routing.Route),
	}

	// Hardcode some values into the store
	store.redirects["/ons"] = routing.Redirect{"/ons", "https://www.ons.gov.uk/", http.StatusTemporaryRedirect}
	store.routes["/moo"] = routing.Route{"/moo", "http://localhost:30001"}

	return store
}

func (s *Store) RegisterReloadFunc(rf func() error) {
	s.reloadFunc = rf
}

func (s *Store) GetRedirects() []routing.Redirect {
	return slices.Collect(maps.Values(s.redirects))
}

func (s *Store) GetRoutes() []routing.Route {
	return []routing.Route{
		{"/moo", "http://localhost:30001"},
	}
}
