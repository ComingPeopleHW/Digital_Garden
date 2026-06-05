package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ComingPeopleHW/Digital_Garden/backend/internal/api"
)

func main() {
	addr := getenv("HTTP_ADDR", ":8080")
	frontendOrigin := getenv("FRONTEND_ORIGIN", "http://localhost:5173")

	server := &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(frontendOrigin),
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
