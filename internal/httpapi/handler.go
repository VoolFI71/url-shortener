package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"url-shortener/internal/shortener"
)

type Handler struct {
	service       *shortener.Service
	publicBaseURL string
}

type createURLRequest struct {
	URL string `json:"url"`
}

type createURLResponse struct {
	ShortURL string `json:"short_url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func New(service *shortener.Service, publicBaseURL string) http.Handler {
	handler := &Handler{
		service:       service,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/urls", handler.createShortURL)
	mux.HandleFunc("GET /{code}", handler.redirect)
	return mux
}

func (h *Handler) createShortURL(w http.ResponseWriter, r *http.Request) {
	var request createURLRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "request body must contain a valid JSON object")
		return
	}

	link, created, err := h.service.Shorten(r.Context(), request.URL)
	if errors.Is(err, shortener.ErrInvalidURL) {
		writeError(w, http.StatusBadRequest, "url must be an absolute HTTP or HTTPS URL")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not shorten URL")
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, createURLResponse{
		ShortURL: h.publicBaseURL + "/" + link.Code,
	})
}

func (h *Handler) redirect(w http.ResponseWriter, r *http.Request) {
	originalURL, err := h.service.Resolve(r.Context(), r.PathValue("code"))
	if errors.Is(err, shortener.ErrNotFound) {
		writeError(w, http.StatusNotFound, "short URL not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not resolve short URL")
		return
	}
	http.Redirect(w, r, originalURL, http.StatusFound)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}
