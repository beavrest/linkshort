package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beavrest/linkshort/internal/model"
	"github.com/beavrest/linkshort/internal/service"
	"github.com/beavrest/linkshort/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandle_PostShorten(t *testing.T) {
	store := storage.NewMemory()
	serviceShortener := service.NewShortenerService(store)
	h := New(serviceShortener, "http://localhost:8080")

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

			h.Shorten(rec, req)

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
	serviceShortener := service.NewShortenerService(store)
	h := New(serviceShortener, "http://localhost:8080")

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
			name:       "GET / (empty id) returns 404",
			path:       "/",
			wantStatus: http.StatusNotFound,
			wantLoc:    "",
		},
	}

	r := chi.NewRouter()
	r.Get("/{id}", h.Expand)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code, "status code")
			if tt.wantLoc != "" {
				assert.Equal(t, tt.wantLoc, rec.Header().Get("Location"), "Location header")
			}
		})
	}
}

func TestHandle_MethodNotAllowed(t *testing.T) {
	store := storage.NewMemory()
	serviceShortener := service.NewShortenerService(store)
	h := New(serviceShortener, "")

	tests := []struct {
		method string
		path   string
	}{
		{"PUT", "/"},
		{"DELETE", "/"},
		{"PATCH", "/xyz"},
		{"POST", "/abc123"},
		{"GET", "/api/shorten"},
		{"PUT", "/api/shorten"},
		{"GET", "/api/shorten/batch"},
		{"PUT", "/api/shorten/batch"},
	}

	r := chi.NewRouter()
	r.Post("/", h.Shorten)
	r.Get("/{id}", h.Expand)
	r.Post("/api/shorten", h.ShortenJSON)
	r.Post("/api/shorten/batch", h.ShortenBatch)

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, "status code")
		})
	}
}

func TestHandle_ShortenUsesBaseURLWhenEmpty(t *testing.T) {
	store := storage.NewMemory()
	serviceShortener := service.NewShortenerService(store)
	h := New(serviceShortener, "")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://ya.ru"))
	req.Host = "localhost:8080"
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	h.Shorten(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, "status")
	body := strings.TrimSpace(rec.Body.String())
	assert.True(t, strings.HasPrefix(body, "http://localhost:8080/"), "body should start with http://localhost:8080/: %s", body)
}

func TestHandle_PostShortenJSON(t *testing.T) {
	store := storage.NewMemory()
	serviceShortener := service.NewShortenerService(store)
	h := New(serviceShortener, "http://localhost:8080")
	tests := []struct {
		name       string
		body       string
		wantStatus int
		checkBody  func(t *testing.T, body []byte)
	}{
		{
			name:       "valid URL returns 201 and JSON result",
			body:       `{"url":"https://practicum.yandex.ru/"}`,
			wantStatus: http.StatusCreated,
			checkBody: func(t *testing.T, body []byte) {
				var resp model.ShortenJSONResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.True(t, strings.HasPrefix(resp.Result, "http://localhost:8080/"), "result: %s", resp.Result)
				suffix := strings.TrimPrefix(resp.Result, "http://localhost:8080/")
				assert.Len(t, suffix, 6, "short ID should be 6 chars")
			},
		},
		{
			name:       "URL with spaces trimmed",
			body:       `{"url":"  https://example.com  "}`,
			wantStatus: http.StatusCreated,
			checkBody: func(t *testing.T, body []byte) {
				var resp model.ShortenJSONResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.True(t, strings.HasPrefix(resp.Result, "http://localhost:8080/"), "result: %s", resp.Result)
			},
		},
		{
			name:       "invalid JSON returns 400",
			body:       `{not json`,
			wantStatus: http.StatusBadRequest,
			checkBody:  nil,
		},
		{
			name:       "empty url returns 400",
			body:       `{"url":""}`,
			wantStatus: http.StatusBadRequest,
			checkBody:  nil,
		},
		{
			name:       "only whitespace in url returns 400",
			body:       `{"url":"   "}`,
			wantStatus: http.StatusBadRequest,
			checkBody:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.ShortenJSON(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code, "status code")
			if tt.wantStatus == http.StatusCreated {
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), "Content-Type")
			}
			if tt.checkBody != nil && tt.wantStatus == http.StatusCreated {
				tt.checkBody(t, rec.Body.Bytes())
			}
		})
	}
}

func TestHandle_ShortenJSONUsesBaseURLWhenEmpty(t *testing.T) {
	store := storage.NewMemory()
	serviceShortener := service.NewShortenerService(store)
	h := New(serviceShortener, "")
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://ya.ru"}`))
	req.Host = "localhost:8080"
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ShortenJSON(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code, "status")
	var resp model.ShortenJSONResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, strings.HasPrefix(resp.Result, "http://localhost:8080/"), "result: %s", resp.Result)
}

func TestHandle_PostShortenConflict(t *testing.T) {
	store := storage.NewMemory()
	serviceShortener := service.NewShortenerService(store)
	h := New(serviceShortener, "http://localhost:8080")

	req1 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/"))
	req1.Header.Set("Content-Type", "text/plain")
	rec1 := httptest.NewRecorder()
	h.Shorten(rec1, req1)
	require.Equal(t, http.StatusCreated, rec1.Code, "first status")
	firstBody := strings.TrimSpace(rec1.Body.String())

	req2 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/"))
	req2.Header.Set("Content-Type", "text/plain")
	rec2 := httptest.NewRecorder()
	h.Shorten(rec2, req2)
	assert.Equal(t, http.StatusConflict, rec2.Code, "second status")
	secondBody := strings.TrimSpace(rec2.Body.String())
	assert.Equal(t, firstBody, secondBody, "conflict response should return the existing short URL")
}

func TestHandle_PostShortenJSONConflict(t *testing.T) {
	store := storage.NewMemory()
	serviceShortener := service.NewShortenerService(store)
	h := New(serviceShortener, "http://localhost:8080")

	body := `{"url":"https://practicum.yandex.ru/"}`

	req1 := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	h.ShortenJSON(rec1, req1)
	require.Equal(t, http.StatusCreated, rec1.Code, "first status")
	var firstResp model.ShortenJSONResponse
	require.NoError(t, json.Unmarshal(rec1.Body.Bytes(), &firstResp))

	req2 := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	h.ShortenJSON(rec2, req2)
	assert.Equal(t, http.StatusConflict, rec2.Code, "second status")
	var secondResp model.ShortenJSONResponse
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &secondResp))
	assert.Equal(t, firstResp.Result, secondResp.Result, "conflict response should return the existing short URL")
}

func TestHandle_PostShortenBatch(t *testing.T) {
	store := storage.NewMemory()
	serviceShortener := service.NewShortenerService(store)
	h := New(serviceShortener, "http://localhost:8080")

	tests := []struct {
		name       string
		body       string
		wantStatus int
		checkBody  func(t *testing.T, body []byte)
	}{
		{
			name:       "valid batch returns 201 with correlation IDs in order",
			body:       `[{"correlation_id":"1","original_url":"https://a.example"},{"correlation_id":"2","original_url":"https://b.example"}]`,
			wantStatus: http.StatusCreated,
			checkBody: func(t *testing.T, body []byte) {
				var resp []model.ShortenBatchResponseItem
				require.NoError(t, json.Unmarshal(body, &resp))
				require.Len(t, resp, 2)
				assert.Equal(t, "1", resp[0].CorrelationID)
				assert.Equal(t, "2", resp[1].CorrelationID)
				assert.True(t, strings.HasPrefix(resp[0].ShortURL, "http://localhost:8080/"), "short_url: %s", resp[0].ShortURL)
				assert.True(t, strings.HasPrefix(resp[1].ShortURL, "http://localhost:8080/"), "short_url: %s", resp[1].ShortURL)
			},
		},
		{
			name:       "empty array returns 400",
			body:       `[]`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON returns 400",
			body:       `[{not json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "item with empty original_url returns 400",
			body:       `[{"correlation_id":"1","original_url":""}]`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.ShortenBatch(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code, "status code")
			if tt.wantStatus == http.StatusCreated {
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), "Content-Type")
			}
			if tt.checkBody != nil && tt.wantStatus == http.StatusCreated {
				tt.checkBody(t, rec.Body.Bytes())
			}
		})
	}
}
