package helpers

import (
	"os"
)

type Config struct {
	UserAddr       string
	CollectionAddr string
	DocumentAddr   string
	SearchAddr     string
	AnalyticsAddr  string
}

func LoadConfig() *Config {
	return &Config{
		UserAddr:       getEnv("USER_SERVICE_URL", "http://localhost:8086"),
		CollectionAddr: getEnv("COLLECTION_SERVICE_URL", "http://localhost:8082"),
		DocumentAddr:   getEnv("DOCUMENT_SERVICE_URL", "http://localhost:8081"),
		SearchAddr:     getEnv("SEARCH_SERVICE_URL", "http://localhost:8084"),
		AnalyticsAddr:  getEnv("ANALYTICS_SERVICE_URL", "http://localhost:8087"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
