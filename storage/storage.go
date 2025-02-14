package storage

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/ONSdigital/dis-routing-go-poc/routing"
)

type Store struct {
	handler    http.Handler
	reloadFunc func() error
}

func NewStore() *Store {
	store := &Store{
		handler: adminHandler(),
	}
	go func() {
		done := false
		for !done {
			time.Sleep(20 * time.Second)
			if store.reloadFunc != nil {
				slog.Debug("triggered reload of router")
				if err := store.reloadFunc(); err != nil {
					slog.Error("unable to reload router", "err", err.Error())
				} else {
					done = true
				}
			}
		}
	}()
	return store
}

func (s *Store) RegisterReloadFunc(rf func() error) {
	s.reloadFunc = rf
}

func (s *Store) GetRedirects() []routing.Redirect {
	return []routing.Redirect{
		{"/ons", "https://www.ons.gov.uk/", http.StatusTemporaryRedirect},
	}
}

func (s *Store) GetRoutes() []routing.Route {
	return []routing.Route{
		{"/moo", "http://localhost:30001"},
	}
}
