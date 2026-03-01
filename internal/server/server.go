package server

import (
	"log"
	"net/http"
)

func Run(addr string, handler http.Handler) error {
	log.Printf("Starting HTTP server on %s", addr)
	return http.ListenAndServe(addr, handler)
}
