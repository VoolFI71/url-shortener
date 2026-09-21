package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VoolFI71/url-shortener/internal/httpapi"
	"github.com/VoolFI71/url-shortener/internal/shortener"
	"github.com/VoolFI71/url-shortener/internal/storage/memory"
	"github.com/VoolFI71/url-shortener/internal/storage/postgres"
)

func main() {
	address := envOrDefault("ADDR", ":8080")
	publicBaseURL := envOrDefault("PUBLIC_BASE_URL", "http://localhost:8080")
	storageType := envOrDefault("STORAGE", "memory")

	connectionContext, cancelConnection := context.WithTimeout(context.Background(), 10*time.Second)
	store, closeStore, err := buildStore(connectionContext, storageType)
	cancelConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer closeStore()

	service := shortener.New(store, shortener.RandomGenerator{})
	server := &http.Server{
		Addr:              address,
		Handler:           httpapi.New(service, publicBaseURL),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("URL shortener is listening on %s", address)
		serverErrors <- server.ListenAndServe()
	}()

	stopContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-stopContext.Done():
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownContext); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

func buildStore(ctx context.Context, storageType string) (shortener.Store, func(), error) {
	switch storageType {
	case "memory":
		return memory.New(), func() {}, nil
	case "postgres":
		databaseURL := os.Getenv("DATABASE_URL")
		if databaseURL == "" {
			return nil, nil, errors.New("DATABASE_URL is required for postgres storage")
		}
		store, err := postgres.Open(ctx, databaseURL)
		if err != nil {
			return nil, nil, err
		}
		return store, store.Close, nil
	default:
		return nil, nil, fmt.Errorf("unsupported STORAGE value %q", storageType)
	}
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
