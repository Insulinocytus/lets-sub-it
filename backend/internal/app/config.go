package app

import "os"

type Config struct {
	Addr    string
	DBPath  string
	WorkDir string
}

func LoadConfig() Config {
	return Config{
		Addr:    envOrDefault("LSI_ADDR", "127.0.0.1:8080"),
		DBPath:  envOrDefault("LSI_DB_PATH", "./data/backend.sqlite3"),
		WorkDir: envOrDefault("LSI_WORK_DIR", "./data/jobs"),
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
