package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/beavrest/linkshort/internal/shortener"
	"github.com/beavrest/linkshort/internal/storage"
)

type Handler struct {
	store   *storage.Memory
	baseURL string
}

func New(store *storage.Memory, baseURL string) *Handler {
	return &Handler{store: store, baseURL: baseURL}
}

// Handle маршрутизирует запросы: POST / - сокращение, GET /{id} — редирект
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/":
		h.shorten(w, r)
	case r.Method == http.MethodGet && r.URL.Path != "/":
		h.expand(w, r)
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) expand(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")
	originalURL, ok := h.store.Get(id)
	if !ok {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
