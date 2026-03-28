package main

import (
	"log"

	"github.com/beavrest/linkshort/internal/config"
	"github.com/beavrest/linkshort/internal/handler"
	"github.com/beavrest/linkshort/internal/logger"
	"github.com/beavrest/linkshort/internal/server"
	"github.com/beavrest/linkshort/internal/service"
	"github.com/beavrest/linkshort/internal/storage"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
)

func main() {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer zapLog.Sync()

	cfg := config.Load()

	store := storage.NewMemory()
	serviceShortener := service.NewShortenerService(store)
	h := handler.New(serviceShortener, cfg.BaseURL)

	r := chi.NewRouter()
	r.Use(logger.WithLogging(zapLog))
	r.Post("/", h.Shorten)
	r.Get("/{id}", h.Expand)

	if err := server.Run(cfg.Addr, r); err != nil {
		log.Fatal(err)
	}
}
