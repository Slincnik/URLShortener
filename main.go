package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

const MAX_ATTEMPTS_CREATE_UNIQ_KEY = 5

type App struct {
	DB *sql.DB
}

type CreateRequest struct {
	URL string `json:"url"`
}

type CreateResponse struct {
	ShortUrl string `json:"short_url"`
}

func NewApp() (*App, error) {
	db, err := sql.Open("sqlite", "file:urls.db?_fk=1")

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			original_url TEXT NOT NULL UNIQUE,
			short_key TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_short_key ON urls(short_key);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON urls(original_url);
		CREATE INDEX IF NOT EXISTS idx_original_url ON urls(original_url);
	`)

	if err != nil {
		db.Close()
		return nil, err
	}
	return &App{DB: db}, nil
}

func generateShortKey() string {
	uuid := uuid.New()
	hash := sha256.Sum256(uuid[:])
	return base64.RawURLEncoding.EncodeToString(hash[:8])
}

func (a *App) createShortURL(originalURL string) (string, error) {
	var existingShortKey string

	err := a.DB.QueryRow("SELECT short_key FROM urls WHERE original_url = ?", originalURL).Scan(&existingShortKey)

	if err == nil && existingShortKey != "" {
		return existingShortKey, nil
	}

	for i := 0; i < MAX_ATTEMPTS_CREATE_UNIQ_KEY; i++ {
		shortKey := generateShortKey()
		result, errInsert := a.DB.Exec("INSERT INTO urls (original_url, short_key) VALUES (?, ?)", originalURL, shortKey)

		if errInsert == nil {
			lastInsertId, _ := result.LastInsertId()
			if lastInsertId > 0 {
				return shortKey, nil
			}
		}

		if strings.Contains(errInsert.Error(), "UNIQUE constraint failed") {
			continue // Попробовать другой ключ
		}

		return "", errInsert
	}

	return "", errors.New("Failed to generate unique short key after multiple attempts")
}

func (a *App) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Проверка валидности URL
	_, err := url.ParseRequestURI(req.URL)
	if err != nil {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	shortKey, err := a.createShortURL(req.URL)

	if err != nil {
		http.Error(w, "Failed to create short URL", http.StatusBadRequest)
		return
	}

	log.Printf("Creating short URL for %s", req.URL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateResponse{ShortUrl: shortKey})
}

func (a *App) handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortKey := strings.TrimPrefix(r.URL.Path, "/")

	if shortKey == "" {
		http.NotFound(w, r)
		return
	}

	var originalURL string

	err := a.DB.QueryRow("SELECT original_url FROM urls WHERE short_key = ?", shortKey).Scan(&originalURL)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	log.Printf("Redirecting %s to %s", shortKey, originalURL)

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func main() {
	app, err := NewApp()

	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}
	defer app.DB.Close()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	mux := http.NewServeMux()

	mux.HandleFunc("/shorten", app.handleCreate)
	mux.HandleFunc("/", app.handleRedirect)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, os.Kill)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	log.Println("Server started on :8080")

	<-done

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server stopped")
}
