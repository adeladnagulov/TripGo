package main

import (
	"context"
	"log"

	"github.com/adeladnagulov/TripGo/internal/config"
	"github.com/adeladnagulov/TripGo/internal/storage"
	"github.com/adeladnagulov/TripGo/internal/usecase"
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
	_ = usecase.NewTripServise(tripRepo, tm)
}
