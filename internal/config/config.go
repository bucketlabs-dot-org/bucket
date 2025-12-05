package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDSN      string
	B2Bucket   string
	B2Endpoint string
	B2Key      string
	B2Secret   string
	BaseURL    string
	Port       string
}

func Load() *Config {
	_ = godotenv.Load() // optional .env

	cfg := &Config{
		DBDSN:      getEnv("DB_DSN", "postgres://bucket:bucketpass@localhost:5432/bucket?sslmode=disable"),
		B2Bucket:   getEnv("B2_BUCKET", ""),
		B2Endpoint: getEnv("B2_ENDPOINT", ""),
		B2Key:      getEnv("B2_ACCESS_KEY", ""),
		B2Secret:   getEnv("B2_SECRET_KEY", ""),
		BaseURL:    getEnv("BASE_URL", "http://localhost:8080"),
		Port:       getEnv("PORT", "8080"),
	}

	if cfg.DBDSN == "" {
		log.Fatal("DB_DSN is required")
	}

	return cfg
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

