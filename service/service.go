package service

import (
	"fmt"

	"github.com/ONSdigital/dis-routing-go-poc/routing"
	"github.com/ONSdigital/dis-routing-go-poc/storage"
	"github.com/ONSdigital/dis-routing-go-poc/upstream"
)

type Service struct {
	RouterPort     int
	UpstreamPort   int
	AdminPort      int
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
		Addr:    fmt.Sprintf("localhost:%d", s.RouterPort),
		Handler: router,
	}
	s.upstreamServer = Server{
		Name:    "upstreamServer",
		Addr:    fmt.Sprintf("localhost:%d", s.UpstreamPort),
		Handler: upstream.Handler(),
	}
	s.storageServer = Server{
		Name:    "storageServer-api",
		Addr:    fmt.Sprintf("localhost:%d", s.AdminPort),
		Handler: store.AdminHandler(),
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
