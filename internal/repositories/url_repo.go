package repositories

import (
	"database/sql"
	"log"
	"urlshortener/config"

	_ "modernc.org/sqlite"
)

type URLRepo interface {
	GetShortKeyByURL(originalURL string) (string, error)
	CreateShortURL(originalURL string, shortKey string) error
	GetOriginalURL(shortKey string) (string, error)
	Close() error
}

type SQLiteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(cfg *config.Config) *SQLiteRepo {
	db, err := sql.Open("sqlite", cfg.DBPath)

	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	db.SetMaxOpenConns(cfg.MaxDBConns)
	db.SetMaxIdleConns(cfg.IdleDBConns)

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	return &SQLiteRepo{db: db}
}

func (r *SQLiteRepo) Migrate() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			original_url TEXT NOT NULL UNIQUE,
			short_key TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_short_key ON urls(short_key);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON urls(original_url);
		CREATE INDEX IF NOT EXISTS idx_original_url ON urls(original_url);;
	`)

	return err
}

func (r *SQLiteRepo) GetShortKeyByURL(originalURL string) (string, error) {
	var existingShortKey string

	err := r.db.QueryRow("SELECT short_key FROM urls WHERE original_url = ?", originalURL).Scan(&existingShortKey)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Failed to get short key by URL: %v", err)
		return "", err
	}

	return existingShortKey, nil
}

func (r *SQLiteRepo) CreateShortURL(originalURL string, shortKey string) error {
	_, err := r.db.Exec("INSERT INTO urls (original_url, short_key) VALUES (?, ?)", originalURL, shortKey)

	if err != nil {
		log.Printf("Failed to create short key by URL: %v", err)
		return err
	}

	return err
}

func (r *SQLiteRepo) GetOriginalURL(shortKey string) (string, error) {
	var originalURL string

	err := r.db.QueryRow("SELECT original_url FROM urls WHERE short_key = ?", shortKey).Scan(&originalURL)

	if err != nil || err != sql.ErrNoRows {
		log.Printf("Failed to get original URL by short key: %v", err)
		return "", err
	}

	return originalURL, nil
}

func (r *SQLiteRepo) Close() error {
	return r.db.Close()
}
