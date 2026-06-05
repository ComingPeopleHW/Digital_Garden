package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ComingPeopleHW/Digital_Garden/backend/internal/api"
	"github.com/ComingPeopleHW/Digital_Garden/backend/internal/store/postgres"
)

func main() {
	addr := getenv("HTTP_ADDR", ":8080")
	frontendOrigin := getenv("FRONTEND_ORIGIN", "http://localhost:5173")
	databaseURL := getenv("DATABASE_URL", "postgres://digital_garden:digital_garden@localhost:5432/digital_garden?sslmode=disable")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		log.Fatalf("database unavailable: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	if err := store.Migrate(ctx); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(frontendOrigin, store),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("digital garden api listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
