package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dis-routing-go-poc/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	slog.Info("Starting POC")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	svc := service.Service{}
	svc.Run()

	select {
	case sig := <-sigChan:
		slog.Info("os signal received", slog.Any("signal", sig))
		svc.Shutdown()
	}
}
