package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/config"
	"github.com/adeladnagulov/TripGo/internal/handler"
	"github.com/adeladnagulov/TripGo/internal/storage"
	"github.com/adeladnagulov/TripGo/internal/usecase"
	"github.com/go-chi/chi/v5"
)

func main() { //убрать в init()
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	connectCtx, cancel := context.WithTimeout(context.Background(), cfg.DatabaseConnectTimeout)
	defer cancel()

	db, err := storage.NewDB(connectCtx, cfg)
	if err != nil {
		log.Fatalf("Failed to load storage: %v", err)
	}
	defer db.Close()

	tripRepo := storage.NewTripRepository(db.Pool, cfg.DatabaseQueryTimeout)
	tm := storage.NewTransactionManager(db.Pool)
	tripService := usecase.NewTripService(tripRepo, tm)

	si := handler.NewHandler(tripService)

	r := chi.NewRouter()

	api.HandlerFromMux(si, r)
	srv := &http.Server{
		Addr:              cfg.HttpAddr,
		Handler:           r,
		ReadTimeout:       cfg.HttpReadTimeout,
		ReadHeaderTimeout: cfg.HttpReadHeaderTimeout,
		WriteTimeout:      cfg.HttpWriteTimeout,
		IdleTimeout:       cfg.HttpIdleTimeout,
	}

	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gracefully")

	shotdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shotdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited gracefully")
}
