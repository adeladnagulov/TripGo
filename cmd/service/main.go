package main

import (
	"context"
	_ "embed"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/config"
	handler "github.com/adeladnagulov/TripGo/internal/http"
	storage "github.com/adeladnagulov/TripGo/internal/repository"

	"github.com/adeladnagulov/TripGo/internal/usecase"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
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
	tm := storage.NewTransactionManager(db.Pool, cfg.DatabaseQueryTimeout)
	tripService := usecase.NewTripService(tripRepo, tm)

	si := handler.NewHandler(tripService)

	specPath := "contracts/openapi/trip-service.openapi.yaml"
	specData, err := os.ReadFile(specPath)
	if err != nil {
		log.Fatalf("Failed to read openapi spec from %s: %v", specPath, err)
	}
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData(specData)
	if err != nil {
		log.Fatalf("Failed to load swagger spec: %v", err)
	}
	if err := swagger.Validate(loader.Context); err != nil {
		log.Fatalf("Swagger spec is invalid: %v", err)
	}
	swagger.Servers = nil

	r := chi.NewRouter()
	r.Use(middleware.OapiRequestValidator(swagger))
	api.HandlerWithOptions(si, api.ChiServerOptions{
		BaseRouter:       r,
		ErrorHandlerFunc: si.HandlerAPIError,
	})
	srv := &http.Server{
		Addr:              cfg.HttpAddr,
		Handler:           r,
		ReadTimeout:       cfg.HttpReadTimeout,
		ReadHeaderTimeout: cfg.HttpReadHeaderTimeout,
		WriteTimeout:      cfg.HttpWriteTimeout,
		IdleTimeout:       cfg.HttpIdleTimeout,
	}

	srvErrCh := make(chan error, 1)
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			srvErrCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("shutting down gracefully")
	case err := <-srvErrCh:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server failed to start or crashed: %v", err)
		}
	}

	shotdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shotdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited gracefully")
}
