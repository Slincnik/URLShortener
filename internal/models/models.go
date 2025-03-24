package models

type CreateRequest struct {
	URL string `json:"url"`
}

type CreateResponse struct {
	ShortURL string `json:"short_url"`
}
