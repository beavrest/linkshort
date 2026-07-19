package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/beavrest/linkshort/internal/config"
	"github.com/beavrest/linkshort/internal/config/db"
	"github.com/beavrest/linkshort/internal/handler"
	"github.com/beavrest/linkshort/internal/logger"
	"github.com/beavrest/linkshort/internal/middleware"
	"github.com/beavrest/linkshort/internal/server"
	"github.com/beavrest/linkshort/internal/service"
	"github.com/beavrest/linkshort/internal/storage"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
)

func main() {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer zapLog.Sync()

	cfg := config.Load()

	database, err := db.NewPostgres(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	var store service.Store
	if cfg.FileStoragePath == "" {
		store = storage.NewMemory()
	} else {
		store, err = storage.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	}

	serviceShortener := service.NewShortenerService(store)
	h := handler.New(serviceShortener, cfg.BaseURL)

	r := chi.NewRouter()
	r.Use(logger.WithLogging(zapLog))
	r.Use(middleware.DecompressRequest)
	r.Use(middleware.CompressResponse)
	r.Post("/", h.Shorten)
	r.Get("/{id}", h.Expand)
	r.Post("/api/shorten", h.ShortenJSON)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if err := database.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	if err := server.Run(cfg.Addr, r); err != nil {
		log.Fatal(err)
	}
}
