package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr    string
	BaseURL string
}

func Load() *Config {
	addr := flag.String("a", "localhost:8080", "HTTP server address")
	baseURL := flag.String("b", "http://localhost:8080", "base URL for short links")

	flag.Parse()

	return &Config{
		Addr:    *addr,
		BaseURL: *baseURL,
	}
}

func GetEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
