package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/beavrest/linkshort/internal/model"

	"github.com/go-chi/chi/v5"
)

type Shortener interface {
	Shorten(originalURL string) (string, error)
	Expand(id string) (string, bool)
}

type Handler struct {
	service Shortener
	baseURL string
}

func New(service Shortener, baseURL string) *Handler {
	return &Handler{service: service, baseURL: baseURL}
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	shortID, err := h.service.Shorten(originalURL)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	base := h.baseURL
	if base == "" {
		base = fmt.Sprintf("http://%s", r.Host)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", base, shortID)
}

func (h *Handler) Expand(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	originalURL, ok := h.service.Expand(id)
	if !ok {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (h *Handler) ShortenJSON(w http.ResponseWriter, r *http.Request) {
	var req model.ShortenJSONRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	originalURL := strings.TrimSpace(req.URL)
	if originalURL == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	shortID, err := h.service.Shorten(originalURL)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	base := h.baseURL
	if base == "" {
		base = fmt.Sprintf("http://%s", r.Host)
	}
	full := fmt.Sprintf("%s/%s", base, shortID)

	resp, err := json.MarshalIndent(model.ShortenJSONResponse{Result: full}, "", "   ")
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(resp)
}
