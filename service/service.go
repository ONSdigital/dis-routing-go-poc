package service

import (
	"fmt"

	"github.com/ONSdigital/dis-routing-go-poc/routing"
	"github.com/ONSdigital/dis-routing-go-poc/storage"
	"github.com/ONSdigital/dis-routing-go-poc/upstream"
)

type Service struct {
	upstreamServer Server
	routingServer  Server
	storageServer  Server
}

func (s *Service) Run() error {

	store := storage.NewStore()
	router, err := routing.NewRouter(store)
	if err != nil {
		return fmt.Errorf("creating router failed: %w", err)
	}

	s.routingServer = Server{
		Name:    "routingServer",
		Addr:    "localhost:30000",
		Handler: router,
	}
	s.upstreamServer = Server{
		Name:    "upstreamServer",
		Addr:    "localhost:30001",
		Handler: upstream.Handler(),
	}
	s.storageServer = Server{
		Name:    "storageServer-api",
		Addr:    "localhost:30002",
		Handler: store,
	}

	s.routingServer.Start()
	s.upstreamServer.Start()
	s.storageServer.Start()

	return nil
}

func (s *Service) Shutdown() {
	s.routingServer.Stop()
	s.upstreamServer.Stop()
	s.storageServer.Stop()
}
