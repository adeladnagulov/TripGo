package container

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/config"
	server "github.com/adeladnagulov/TripGo/internal/http"
	"github.com/adeladnagulov/TripGo/internal/http/health"
	"github.com/adeladnagulov/TripGo/internal/http/helper/response"
	"github.com/adeladnagulov/TripGo/internal/http/middleware"
	"github.com/adeladnagulov/TripGo/internal/http/trip/create"
	"github.com/adeladnagulov/TripGo/internal/http/trip/finish"
	"github.com/adeladnagulov/TripGo/internal/http/trip/get"
	"github.com/adeladnagulov/TripGo/internal/infra/postgres"
	txManager "github.com/adeladnagulov/TripGo/internal/infra/postgres/transaction"
	"github.com/adeladnagulov/TripGo/internal/infra/telemetry"
	"github.com/adeladnagulov/TripGo/internal/repository/trip"
	createUsecase "github.com/adeladnagulov/TripGo/internal/usecase/trip/create"
	finishUsecase "github.com/adeladnagulov/TripGo/internal/usecase/trip/finish"
	getUsecase "github.com/adeladnagulov/TripGo/internal/usecase/trip/get"
)

type Container struct {
	srv    *http.Server
	cfg    *config.Config
	db     *postgres.DB
	tracer *telemetry.OTelProviders
}

func New(cfg *config.Config) (*Container, error) {
	connectCtx, cancel := context.WithTimeout(context.Background(), cfg.DatabaseConnectTimeout)
	defer cancel()

	provider, err := telemetry.NewOTel(connectCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init tracer: %w", err)
	}
	logger := telemetry.NewLogger(config.ParseLodLevel(cfg.LogLevel))
	meter := provider.Meter.Meter(cfg.OtelServiceName)

	db, err := postgres.NewDB(connectCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load storage: %v", err)
	}

	tripRepo := trip.NewTripRepository(db.Pool, cfg.DatabaseQueryTimeout)
	tm := txManager.NewTransactionManager(db.Pool, cfg.DatabaseQueryTimeout)

	createUc := createUsecase.NewCreateUseCase(tripRepo, tm, logger, meter)
	finishUc := finishUsecase.NewFinishUseCase(tripRepo, tm, logger, meter)
	getUc := getUsecase.NewGetUseCase(tripRepo, logger)

	createH := create.NewHandler(createUc)
	finishH := finish.NewHandler(finishUc)
	getH := get.NewHandler(getUc)
	healthH := health.NewHandler(db.Pool)
	si := server.NewServerInterface(getH, createH, finishH, healthH)

	specPath := "contracts/openapi/trip-service.openapi.yaml"
	specData, err := os.ReadFile(specPath)
	if err != nil {
		log.Printf("Failed to read openapi spec from %s: %v", specPath, err)
	}
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData(specData)
	if err != nil {
		log.Printf("Failed to load swagger spec: %v", err)
	}
	if err := swagger.Validate(loader.Context); err != nil {
		log.Printf("Swagger spec is invalid: %v", err)
	}
	swagger.Servers = nil

	r := chi.NewRouter()
	openAPIRequestValidator, err := middleware.OpenAPIRequestValidator(swagger)
	if err != nil {
		return nil, fmt.Errorf("failed to init OpenAPI validator: %w", err)
	}
	r.Use(openAPIRequestValidator)
	r.Use(middleware.Recover(logger))
	r.Use(middleware.Logging(logger))
	api.HandlerWithOptions(si, api.ChiServerOptions{
		BaseRouter: r,
		Middlewares: []api.MiddlewareFunc{
			api.MiddlewareFunc(middleware.Tracing()),
			api.MiddlewareFunc(middleware.Metrics(meter)),
		},
		ErrorHandlerFunc: response.HandleAPIError,
	})

	srv := &http.Server{
		Addr:              cfg.HttpAddr,
		Handler:           r,
		ReadTimeout:       cfg.HttpReadTimeout,
		ReadHeaderTimeout: cfg.HttpReadHeaderTimeout,
		WriteTimeout:      cfg.HttpWriteTimeout,
		IdleTimeout:       cfg.HttpIdleTimeout,
	}
	return &Container{srv: srv, cfg: cfg, db: db, tracer: provider}, nil
}

func (c *Container) Run() {
	defer c.db.Close()
	srvErrCh := make(chan error, 1)
	go func() {
		log.Printf("Starting server on %s", c.srv.Addr)
		if err := c.srv.ListenAndServe(); err != nil {
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

	shotdownCtx, cancel := context.WithTimeout(context.Background(), c.cfg.ShutdownTimeout)
	defer cancel()

	if c.tracer != nil {
		defer c.tracer.Shutdown(shotdownCtx)
	}

	if err := c.srv.Shutdown(shotdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited gracefully")
}
