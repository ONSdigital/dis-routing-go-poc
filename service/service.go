package service

import (
	"github.com/ONSdigital/dis-routing-go-poc/routing"
	"github.com/ONSdigital/dis-routing-go-poc/storage"
	"github.com/ONSdigital/dis-routing-go-poc/upstream"
)

type Service struct {
	upstream Server
	routing  Server
	storage  Server
}

func (s *Service) Run() {

	s.routing = Server{
		Name:    "routing",
		Addr:    "localhost:30000",
		Handler: routing.Handler{},
	}
	s.upstream = Server{
		Name:    "upstream",
		Addr:    "localhost:30001",
		Handler: upstream.Handler(),
	}
	s.storage = Server{
		Name:    "storage-api",
		Addr:    "localhost:30002",
		Handler: storage.AdminHandler(),
	}

	s.routing.Start()
	s.upstream.Start()
	s.storage.Start()
}

func (s *Service) Shutdown() {
	s.routing.Stop()
	s.upstream.Stop()
	s.storage.Stop()
}
