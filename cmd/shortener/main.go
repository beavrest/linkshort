package main

import (
	"log"
	"net/http"

	handlers "github.com/beavrest/linkshort/internal/handler"
	"github.com/beavrest/linkshort/internal/storage"
)

func main() {
	store := storage.NewMemory()
	h := handlers.New(store, "")

	http.HandleFunc("/", h.Handle)

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
