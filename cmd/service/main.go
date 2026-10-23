package main

import (
	"log"

	"github.com/adeladnagulov/TripGo/internal/config"
	"github.com/adeladnagulov/TripGo/internal/container"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	container, err := container.New(cfg)
	if err != nil {
		log.Fatalf("failed to build container: %v", err)
	}

	container.Run()
}
