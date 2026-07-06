package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr            string
	BaseURL         string
	FileStoragePath string
}

func Load() *Config {
	addr := flag.String("a", "localhost:8080", "HTTP server address")
	baseURL := flag.String("b", "http://localhost:8080", "base URL for short links")
	filePath := flag.String("f", "shortener_storage.json", "file storage path")

	flag.Parse()

	return &Config{
		Addr:            GetEnv("SERVER_ADDRESS", *addr),
		BaseURL:         GetEnv("BASE_URL", *baseURL),
		FileStoragePath: GetEnv("FILE_STORAGE_PATH", *filePath),
	}
}

func GetEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
