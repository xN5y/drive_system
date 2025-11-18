package config

import (
	"fmt"
	"os"
)

type Config struct {
	StorageBackend   string
	BearerToken      string
	LocalStoragePath string
	S3Endpoint       string
	S3AccessKey      string
	S3SecretKey      string
	S3BucketName     string
	S3Region         string
	DatabaseURL      string
	ServerPort       string
}

func Load() (*Config, error) {
	cfg := &Config{
		StorageBackend:   getEnv("STORAGE_BACKEND", "local"),
		BearerToken:      getEnv("BEARER_TOKEN", "(^&%sdfuyigsdfiuhgy(^&*"),
		LocalStoragePath: getEnv("LOCAL_STORAGE_PATH", "./storage_data"),
		S3Endpoint:       getEnv("S3_ENDPOINT", "http://localhost:9000"),
		S3AccessKey:      getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey:      getEnv("S3_SECRET_KEY", "minioadmin"),
		S3BucketName:     getEnv("S3_BUCKET_NAME", "rekaz-bucket"),
		S3Region:         getEnv("S3_REGION", "us-east-1"),
		DatabaseURL:      getEnv("DATABASE_URL", "./metadata.db"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
	}

	if cfg.BearerToken == "" {
		return nil, fmt.Errorf("BEARER_TOKEN is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
