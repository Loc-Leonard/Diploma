package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDsn        string
	AdminEmail   string
	AdminName    string
	AdminPass    string
	JWTSecret    string
	StorageRoot  string
	CVServiceURL string
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		DBDsn:        os.Getenv("DB_DSN"),
		AdminEmail:   os.Getenv("ADMIN_EMAIL"),
		AdminName:    os.Getenv("ADMIN_FULL_NAME"),
		AdminPass:    os.Getenv("ADMIN_PASSWORD"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		StorageRoot:  getenvDefault("STORAGE_ROOT", "storage"),
		CVServiceURL: getenvDefault("CV_SERVICE_URL", "http://cv-agent:8000"),
	}

	if cfg.DBDsn == "" {
		log.Fatal("DB_DSN is required")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	return cfg
}

func getenvDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
