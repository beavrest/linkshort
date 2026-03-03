package config

import (
	"flag"
)

type Config struct {
	Addr    string
	BaseURL string
}

func MustLoad() Config {
	addr := flag.String("a", "localhost:8080", "HTTP server address")
	baseURL := flag.String("b", "http://localhost:8080", "base URL for short links")

	flag.Parse()

	return Config{
		Addr:    *addr,
		BaseURL: *baseURL,
	}
}
