package models

import "time"

type Blob struct {
	ID   string
	Data []byte
}

type BlobResponse struct {
	ID        string `json:"id"`
	Data      string `json:"data"`
	Size      string `json:"size"`
	CreatedAt string `json:"created_at"`
}

type BlobRequest struct {
	ID   string `json:"id" binding:"required"`
	Data string `json:"data" binding:"required"`
}

type Metadata struct {
	ID          string
	Size        int64
	CreatedAt   time.Time
	StorageType string
	StoragePath string
}
