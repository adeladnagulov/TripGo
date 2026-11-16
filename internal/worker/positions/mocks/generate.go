package mocks

//go:generate mockgen -destination=gateway_mock.go -package=mocks github.com/adeladnagulov/TripGo/internal/worker/positions/worker_pool Gateway
//go:generate mockgen -destination=stale_trips_getter_mock.go -package=mocks github.com/adeladnagulov/TripGo/internal/worker/positions/scanner StaleTripsGetter
//go:generate mockgen -destination=ticker_mock.go -package=mocks github.com/adeladnagulov/TripGo/internal/worker/positions/scanner Ticker
