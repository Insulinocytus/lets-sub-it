package config

import "os"

type Config struct {
	DatabasePath string
	StorageDir   string
	BackendAddr  string
}

func Load() Config {
	return Config{
		DatabasePath: getenv("DATABASE_PATH", "./data/app.db"),
		StorageDir:   getenv("STORAGE_DIR", "./data/assets"),
		BackendAddr:  getenv("BACKEND_ADDR", ":8080"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
