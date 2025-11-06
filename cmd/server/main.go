package main

import (
	"log"

	"example.com/avalanche/internal/app"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}
	a.Run()
}
