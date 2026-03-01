package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/beavrest/linkshort/internal/repository"
	"github.com/beavrest/linkshort/internal/shortener"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store   repository.Storer
	baseURL string
}

func New(store repository.Storer, baseURL string) *Handler {
	return &Handler{store: store, baseURL: baseURL}
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

	shortID := shortener.GenerateID()
	h.store.Save(shortID, originalURL)

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
	originalURL, ok := h.store.Get(id)
	if !ok {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
