package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipReadCloser struct {
	io.ReadCloser
	zr *gzip.Reader
}

func (g *gzipReadCloser) Close() error {
	_ = g.zr.Close()
	return g.ReadCloser.Close()
}

func DecompressRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			next.ServeHTTP(w, r)
			return
		}
		zr, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "bad gzip body", http.StatusBadRequest)
			return
		}
		r.Body = &gzipReadCloser{
			ReadCloser: io.NopCloser(zr),
			zr:         zr,
		}
		next.ServeHTTP(w, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gz      *gzip.Writer
	enabled bool
	decided bool
}

func (grw *gzipResponseWriter) decide() {
	if grw.decided {
		return
	}
	grw.decided = true
	ct := grw.Header().Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "text/html") {
		grw.enabled = true
		grw.Header().Set("Content-Encoding", "gzip")
		grw.Header().Add("Vary", "Accept-Encoding")
		grw.Header().Del("Content-Length")
		grw.gz = gzip.NewWriter(grw.ResponseWriter)
	}
}

func (grw *gzipResponseWriter) WriteHeader(statusCode int) {
	grw.decide()
	grw.ResponseWriter.WriteHeader(statusCode)
}

func (grw *gzipResponseWriter) Write(b []byte) (int, error) {
	grw.decide()
	if grw.enabled {
		return grw.gz.Write(b)
	}
	return grw.ResponseWriter.Write(b)
}

func (grw *gzipResponseWriter) Close() error {
	if grw.gz != nil {
		return grw.gz.Close()
	}
	return nil
}

func CompressResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		grw := &gzipResponseWriter{
			ResponseWriter: w,
		}
		defer grw.Close()
		next.ServeHTTP(grw, r)
	})
}
