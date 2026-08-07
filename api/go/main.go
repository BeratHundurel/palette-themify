package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	"themesmith/config"
	"themesmith/db"
	"themesmith/handlers"
	"themesmith/telemetry"

	_ "golang.org/x/image/webp"
)

func main() {
	appConfig, err := config.Load()
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	shutdownTelemetry, err := telemetry.Initialize(context.Background(), telemetry.Config{
		Environment:    appConfig.Environment,
		ServiceName:    appConfig.ServiceName,
		ServiceVersion: appConfig.ServiceVersion,
	})
	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer func() {
		ctx, cancel := telemetry.ShutdownContext()
		defer cancel()
		if err := shutdownTelemetry(ctx); err != nil {
			log.Printf("Failed to flush telemetry: %v", err)
		}
	}()

	if err := db.InitDatabase(); err != nil {
		log.Printf("Failed to initialize database: %v", err)
		log.Println("Continuing without database functionality...")
	}

	defer func() {
		if db.DB != nil {
			if err := db.CloseDatabase(); err != nil {
				log.Printf("Failed to close database: %v", err)
			}
		}
	}()

	router := handlers.NewRouter(appConfig)

	server := &http.Server{
		Addr:              appConfig.HTTPAddress,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	shutdownSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("Starting server on %s (%s)", appConfig.HTTPAddress, appConfig.Environment)
	serverError := make(chan error, 1)
	go func() {
		serverError <- server.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server stopped unexpectedly: %v", err)
		}
	case <-shutdownSignal.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
		}
	}
}
