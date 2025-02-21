package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dis-routing-go-poc/service"
)

func main() {
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(logHandler))
	slog.Info("Starting POC")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	svc := service.Service{
		RouterPort:   30000,
		UpstreamPort: 30001,
		AdminPort:    30002,
	}
	err := svc.Run()
	if err != nil {
		slog.Error("failed to run service", "error", err)
		os.Exit(1)
	}

	select {
	case sig := <-sigChan:
		slog.Info("os signal received", slog.Any("signal", sig))
		svc.Shutdown()
	}
}
