package main

import (
	"fmt"
	"log"

	"simple-drive/internal/api"
	"simple-drive/internal/config"
	"simple-drive/internal/metadata"
	"simple-drive/internal/storage"
	"simple-drive/internal/storage/database"
	"simple-drive/internal/storage/local"
	"simple-drive/internal/storage/s3"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	metadataService, err := metadata.NewService(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize metadata service: %v", err)
	}
	defer metadataService.Close()

	var storageBackend storage.Storage
	switch cfg.StorageBackend {
	case "local":
		storageBackend, err = local.NewLocalStorage(cfg.LocalStoragePath)
		if err != nil {
			log.Fatalf("Failed to initialize local storage: %v", err)
		}
	case "s3":
		storageBackend, err = s3.NewS3Storage(
			cfg.S3Endpoint,
			cfg.S3AccessKey,
			cfg.S3SecretKey,
			cfg.S3BucketName,
			cfg.S3Region,
		)
		if err != nil {
			log.Fatalf("Failed to initialize S3 storage: %v", err)
		}
	case "database":
		dbStorage, err := database.NewDatabaseStorage(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to initialize database storage: %v", err)
		}
		defer dbStorage.Close()
		storageBackend = dbStorage
	default:
		log.Fatalf("Unknown storage backend: %s", cfg.StorageBackend)
	}

	handler := api.NewHandler(storageBackend, metadataService, cfg.StorageBackend)
	router := api.SetupRouter(handler, cfg.BearerToken)

	log.Printf("Starting Simple Drive server on port %s with %s storage backend", cfg.ServerPort, cfg.StorageBackend)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func init() {
	fmt.Println("Simple Drive - Object Storage System")
	fmt.Println("====================================")
}
