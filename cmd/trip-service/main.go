package main

import (
	"log"

	"github.com/adeladnagulov/TripGo/internal/config"
)

func main() {
	_, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
}
