package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/beavrest/linkshort/internal/config"
	"github.com/beavrest/linkshort/internal/config/db"
	"github.com/beavrest/linkshort/internal/handler"
	"github.com/beavrest/linkshort/internal/logger"
	"github.com/beavrest/linkshort/internal/middleware"
	"github.com/beavrest/linkshort/internal/repository"
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

	var database *sql.DB
	var store service.Store

	switch {
	case cfg.DatabaseDSN != "":
		database, err = db.NewPostgres(cfg.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		if err := repository.Migrate(database); err != nil {
			log.Fatal(err)
		}
		store = repository.NewPostgresStore(database)

	case cfg.FileStoragePath != "":
		store, err = storage.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}

	default:
		store = storage.NewMemory()
	}

	if database != nil {
		defer database.Close()
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
		if database == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
