package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"urlshortener/config"
	"urlshortener/internal/handlers"
	"urlshortener/internal/repositories"
	"urlshortener/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.LoadConfig(".")

	r, cleanup := setupRouter(cfg)

	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           r,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	log.Printf("Server started on :8080")

	<-done

	defer cleanup()

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server stopped")
}

func setupRouter(cfg *config.Config) (*chi.Mux, func()) {
	repository := repositories.NewUrlRepo(cfg)

	service := services.NewURLService(repository, cfg)

	r := chi.NewRouter()

	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: log.Default(), NoColor: false}))
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)

	handlers.NewURLHandler(service).RegisterRoutes(r)

	cleanup := func() {
		if err := repository.Close(); err != nil {
			log.Fatalf("Failed to close repository: %v", err)
		}
	}

	return r, cleanup
}
