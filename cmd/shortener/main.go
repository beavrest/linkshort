package main

import (
	"log"

	"github.com/beavrest/linkshort/internal/handler"
	"github.com/beavrest/linkshort/internal/server"
	"github.com/beavrest/linkshort/internal/storage"

	"github.com/go-chi/chi/v5"
)

func main() {
	store := storage.NewMemory()
	h := handler.New(store, "")

	r := chi.NewRouter()
	r.Post("/", h.Shorten)
	r.Get("/{id}", h.Expand)

	if err := server.Run("localhost:8080", r); err != nil {
		log.Fatal(err)
	}
}
