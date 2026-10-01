package main

import (
	"context"
	"log"
	"net/http"

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

	log.Printf("Starting server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
