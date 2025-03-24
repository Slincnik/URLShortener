package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"urlshortener/internal/models"
	"urlshortener/internal/services"

	"github.com/go-chi/chi/v5"
)

type URLHandler struct {
	service *services.URLService
}

func NewURLHandler(service *services.URLService) *URLHandler {
	return &URLHandler{service: service}
}

func (u *URLHandler) RegisterRoutes(r *chi.Mux) {
	r.Post("/shorten", u.HandleCreate)
	r.Get("/{shortKey}", u.HandleRedirect)
	r.Get("/health", u.HandleHealth)
}

func (u *URLHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRequest

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

	shortKey, err := u.service.CreateShortURL(req.URL)

	if err != nil {
		http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err = json.NewEncoder(w).Encode(models.CreateResponse{ShortURL: shortKey}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (u *URLHandler) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	shortKey := chi.URLParam(r, "shortKey")

	if shortKey == "" {
		http.NotFound(w, r)
		return
	}

	originalURL, err := u.service.GetOriginalURL(shortKey)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (u *URLHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
