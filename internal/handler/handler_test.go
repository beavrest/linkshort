package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beavrest/linkshort/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandle_PostShorten(t *testing.T) {
	store := storage.NewMemory()
	h := New(store, "http://localhost:8080")

	tests := []struct {
		name       string
		body       string
		wantStatus int
		checkBody  func(t *testing.T, body string)
	}{
		{
			name:       "valid URL returns 201 and short URL",
			body:       "https://practicum.yandex.ru/",
			wantStatus: http.StatusCreated,
			checkBody: func(t *testing.T, body string) {
				assert.True(t, strings.HasPrefix(body, "http://localhost:8080/"), "body should start with base URL: %s", body)
				suffix := strings.TrimPrefix(body, "http://localhost:8080/")
				assert.Len(t, suffix, 6, "short ID should be 6 chars")
			},
		},
		{
			name:       "URL with spaces trimmed",
			body:       "  https://example.com  ",
			wantStatus: http.StatusCreated,
			checkBody: func(t *testing.T, body string) {
				assert.True(t, strings.HasPrefix(body, "http://localhost:8080/"), "body should start with base URL: %s", body)
			},
		},
		{
			name:       "empty body returns 400",
			body:       "",
			wantStatus: http.StatusBadRequest,
			checkBody:  nil,
		},
		{
			name:       "only whitespace returns 400",
			body:       "   ",
			wantStatus: http.StatusBadRequest,
			checkBody:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "text/plain")
			rec := httptest.NewRecorder()

			h.Handle(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code, "status code")
			if tt.wantStatus == http.StatusCreated {
				assert.Equal(t, "text/plain", rec.Header().Get("Content-Type"), "Content-Type")
			}
			if tt.checkBody != nil && tt.wantStatus == http.StatusCreated {
				tt.checkBody(t, strings.TrimSpace(rec.Body.String()))
			}
		})
	}
}

func TestHandle_GetExpand(t *testing.T) {
	store := storage.NewMemory()
	store.Save("abc123", "https://practicum.yandex.ru/")
	h := New(store, "http://localhost:8080")

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantLoc    string
	}{
		{
			name:       "existing id returns 307 and Location",
			path:       "/abc123",
			wantStatus: http.StatusTemporaryRedirect,
			wantLoc:    "https://practicum.yandex.ru/",
		},
		{
			name:       "non-existing id returns 404",
			path:       "/nonexistent",
			wantStatus: http.StatusNotFound,
			wantLoc:    "",
		},
		{
			name:       "GET / returns 405",
			path:       "/",
			wantStatus: http.StatusMethodNotAllowed,
			wantLoc:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			h.Handle(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code, "status code")
			if tt.wantLoc != "" {
				assert.Equal(t, tt.wantLoc, rec.Header().Get("Location"), "Location header")
			}
		})
	}
}

func TestHandle_MethodNotAllowed(t *testing.T) {
	store := storage.NewMemory()
	h := New(store, "")

	tests := []struct {
		method string
		path   string
	}{
		{"PUT", "/"},
		{"DELETE", "/"},
		{"PATCH", "/xyz"},
		{"POST", "/abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			h.Handle(rec, req)

			assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, "status code")
		})
	}
}

func TestHandle_ShortenUsesBaseURLWhenEmpty(t *testing.T) {
	store := storage.NewMemory()
	h := New(store, "")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://ya.ru"))
	req.Host = "localhost:8080"
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	h.Handle(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, "status")
	body := strings.TrimSpace(rec.Body.String())
	assert.True(t, strings.HasPrefix(body, "http://localhost:8080/"), "body should start with http://localhost:8080/: %s", body)
}
