package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/beavrest/linkshort/internal/model"
	"github.com/beavrest/linkshort/internal/service"

	"github.com/go-chi/chi/v5"
)

type Shortener interface {
	Shorten(originalURL string) (string, error)
	Expand(id string) (string, bool)
	ShortenBatch(originalURLs []string) ([]string, error)
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
	status := http.StatusCreated
	if errors.Is(err, service.ErrURLExists) {
		status = http.StatusConflict
	} else if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	base := h.baseURL
	if base == "" {
		base = fmt.Sprintf("http://%s", r.Host)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	originalURL := strings.TrimSpace(req.URL)
	if originalURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortID, err := h.service.Shorten(originalURL)
	status := http.StatusCreated
	if errors.Is(err, service.ErrURLExists) {
		status = http.StatusConflict
	} else if err != nil {
		log.Printf("shorten_json: service shorten error: %v (url=%q)", err, originalURL)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	base := h.baseURL
	if base == "" {
		base = fmt.Sprintf("http://%s", r.Host)
	}

	full, err := url.JoinPath(base, shortID)
	if err != nil {
		// это внутренняя ошибка сборки URL
		log.Printf("shorten_json: join path error: %v (base=%q, id=%q)", err, base, shortID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.MarshalIndent(model.ShortenJSONResponse{Result: full}, "", "   ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(resp)
}

func (h *Handler) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	var req []model.ShortenBatchRequestItem
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(req) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURLs := make([]string, len(req))
	for i, item := range req {
		u := strings.TrimSpace(item.OriginalURL)
		if u == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		originalURLs[i] = u
	}

	shortIDs, err := h.service.ShortenBatch(originalURLs)
	if err != nil {
		log.Printf("shorten_batch: service shorten batch error: %v (count=%d)", err, len(originalURLs))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	base := h.baseURL
	if base == "" {
		base = fmt.Sprintf("http://%s", r.Host)
	}

	resp := make([]model.ShortenBatchResponseItem, len(req))
	for i, item := range req {
		full, err := url.JoinPath(base, shortIDs[i])
		if err != nil {
			log.Printf("shorten_batch: join path error: %v (base=%q, id=%q)", err, base, shortIDs[i])
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp[i] = model.ShortenBatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      full,
		}
	}

	body, err := json.MarshalIndent(resp, "", "   ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(body)
}
