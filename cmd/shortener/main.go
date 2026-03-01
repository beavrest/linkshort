package main

import (
	"log"
	"net/http"

	"github.com/beavrest/linkshort/internal/handler"
	"github.com/beavrest/linkshort/internal/server"
	"github.com/beavrest/linkshort/internal/storage"
)

func main() {
	store := storage.NewMemory()
	h := handler.New(store, "")

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.Handle)
	if err := server.Run("localhost:8080", mux); err != nil {
		log.Fatal(err)
	}
}
