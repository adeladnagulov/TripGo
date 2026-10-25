package container

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/config"
	server "github.com/adeladnagulov/TripGo/internal/http"
	"github.com/adeladnagulov/TripGo/internal/http/health"
	"github.com/adeladnagulov/TripGo/internal/http/trip/create"
	"github.com/adeladnagulov/TripGo/internal/http/trip/finish"
	"github.com/adeladnagulov/TripGo/internal/http/trip/get"
	"github.com/adeladnagulov/TripGo/internal/infra/postgres"
	txManager "github.com/adeladnagulov/TripGo/internal/infra/postgres/transaction"
	"github.com/adeladnagulov/TripGo/internal/repository/trip"
	createUsecase "github.com/adeladnagulov/TripGo/internal/usecase/trip/create"
	finishUsecase "github.com/adeladnagulov/TripGo/internal/usecase/trip/finish"
	getUsecase "github.com/adeladnagulov/TripGo/internal/usecase/trip/get"
	"github.com/go-chi/chi/v5"
)

type Container struct {
	srv *http.Server
	cfg *config.Config
	db  *postgres.DB
}

func New(cfg *config.Config) (*Container, error) {
	connectCtx, cancel := context.WithTimeout(context.Background(), cfg.DatabaseConnectTimeout)
	defer cancel()

	db, err := postgres.NewDB(connectCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to load storage: %v", err)
	}

	tripRepo := trip.NewTripRepository(db.Pool, cfg.DatabaseQueryTimeout)
	tm := txManager.NewTransactionManager(db.Pool)

	createUc := createUsecase.NewCreateUseCase(tripRepo, tm)
	finishUc := finishUsecase.NewFinishUseCase(tripRepo, tm)
	getUc := getUsecase.NewCreateUseCase(tripRepo)

	createH := create.NewHandler(createUc)
	finishH := finish.NewHandler(finishUc)
	getH := get.NewHandler(getUc)

	healthH := health.NewHandler(db.Pool)

	si := server.NewServerInterface(getH, createH, finishH, healthH)
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
	return &Container{srv: srv, cfg: cfg, db: db}, nil
}

func (c *Container) Run() {
	defer c.db.Close()
	go func() {
		log.Printf("Starting server on %s", c.srv.Addr)
		if err := c.srv.ListenAndServe(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gracefully")

	shotdownCtx, cancel := context.WithTimeout(context.Background(), c.cfg.ShutdownTimeout)
	defer cancel()

	if err := c.srv.Shutdown(shotdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited gracefully")
}
