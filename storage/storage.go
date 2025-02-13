package storage

import (
	"net/http"

	"github.com/ONSdigital/dis-routing-go-poc/routing"
)

type Store struct {
	handler http.Handler
}

func NewStore() *Store {
	store := &Store{
		handler: adminHandler(),
	}
	return store
}

func (s Store) GetRedirects() []routing.Redirect {
	return []routing.Redirect{
		{"/ons", "https://www.ons.gov.uk/", http.StatusTemporaryRedirect},
	}
}
