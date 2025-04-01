package repositories

import (
	"fmt"
	"log"
	"time"
	"urlshortener/config"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type URLRepo interface {
	GetShortKeyByURL(originalURL string) (string, error)
	CreateShortURL(originalURL string, shortKey string) error
	GetOriginalURL(shortKey string) (string, error)
	Close() error
}

type URL struct {
	ID          uint   `gorm:"primary_key"`
	OriginalURL string `gorm:"column:original_url;unique;not null;"`
	ShortKey    string `gorm:"column:short_key;unique;not null;"`
	CreatedAt   time.Time
}

type UrlRepo struct {
	db *gorm.DB
}

func NewUrlRepo(cfg *config.Config) *UrlRepo {
	var dsn string
	switch cfg.DBType {
	case "sqlite":
		dsn = cfg.DBPath
	case "postgres":
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	default:
		log.Fatalf("Unsupported database type: %s", cfg.DBType)
	}

	var db *gorm.DB
	var err error

	logger := logger.New(
		log.New(log.Default().Writer(), "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: false,       // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)

	if cfg.DBType == "sqlite" {
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger,
		})
	} else {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger,
		})
	}

	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	sqlDB, _ := db.DB()

	if err = sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxDBConns)
	sqlDB.SetMaxIdleConns(cfg.IdleDBConns)

	if cfg.Env == config.EnvDev {
		err := db.Migrator().AutoMigrate(&URL{})

		if err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}
	}

	return &UrlRepo{db: db}
}

func (r *UrlRepo) GetShortKeyByURL(originalURL string) (string, error) {
	var shortKey string

	err := r.db.Model(&URL{}).Where("original_url = ?", originalURL).Select("short_key").Scan(&shortKey).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Failed to get short key by URL: %v", err)
	}

	return shortKey, err
}

func (r *UrlRepo) CreateShortURL(originalURL string, shortKey string) error {
	url := URL{
		OriginalURL: originalURL,
		ShortKey:    shortKey,
	}

	err := r.db.Create(&url).Error

	if err != nil {
		log.Printf("Failed to create short URL: %v", err)
	}

	return err
}

func (r *UrlRepo) GetOriginalURL(shortKey string) (string, error) {
	var originalURL string

	err := r.db.Model(&URL{}).Where("short_key = ?", shortKey).Select("original_url").Scan(&originalURL).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Failed to get original URL by short key: %v", err)
	}

	return originalURL, err
}

func (r *UrlRepo) Close() error {
	sqlDB, _ := r.db.DB()

	return sqlDB.Close()
}
