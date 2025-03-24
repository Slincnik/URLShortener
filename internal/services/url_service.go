package services

import (
	"errors"
	"strings"
	"urlshortener/config"
	"urlshortener/internal/repositories"
	"urlshortener/pkg/utils"
)

type URLService struct {
	repo repositories.URLRepo
	cfg  *config.Config
}

func NewURLService(repo repositories.URLRepo, cfg *config.Config) *URLService {
	return &URLService{repo: repo, cfg: cfg}
}

func (s *URLService) CreateShortURL(originalURL string) (string, error) {

	existing, err := s.repo.GetShortKeyByURL(originalURL)

	if err == nil && existing != "" {
		return existing, nil
	}

	for range s.cfg.MaxAttemptsCreateKey {
		shortKey := utils.GenerateShortKey()

		err := s.repo.CreateShortURL(originalURL, shortKey)

		if err == nil {
			return shortKey, nil
		}

		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			continue // Попробовать другой ключ
		}

		return "", err
	}

	return "", errors.New("failed to generate unique short key after multiple attempts")
}

func (s *URLService) GetOriginalURL(shortKey string) (string, error) {
	return s.repo.GetOriginalURL(shortKey)
}
