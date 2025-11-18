package api

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"simple-drive/internal/metadata"
	"simple-drive/internal/models"
	"simple-drive/internal/storage"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	storage         storage.Storage
	metadataService *metadata.Service
	storageType     string
}

func NewHandler(storage storage.Storage, metadataService *metadata.Service, storageType string) *Handler {
	return &Handler{
		storage:         storage,
		metadataService: metadataService,
		storageType:     storageType,
	}
}

// CreateBlob handles POST /v1/blobs
func (h *Handler) CreateBlob(c *gin.Context) {
	var req models.BlobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 data"})
		return
	}

	blob := &models.Blob{
		ID:   req.ID,
		Data: data,
	}

	storagePath, err := h.storage.Save(blob)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save blob: %v", err)})
		return
	}

	metadata := &models.Metadata{
		ID:          req.ID,
		Size:        int64(len(data)),
		CreatedAt:   time.Now().UTC(),
		StorageType: h.storageType,
		StoragePath: storagePath,
	}

	if err := h.metadataService.Save(metadata); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save metadata: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      req.ID,
		"message": "Blob created successfully",
	})
}

// GetBlob handles GET /v1/blobs/:id
func (h *Handler) GetBlob(c *gin.Context) {
	id := c.Param("id")

	metadata, err := h.metadataService.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blob not found"})
		return
	}

	blob, err := h.storage.Retrieve(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to retrieve blob: %v", err)})
		return
	}

	response := models.BlobResponse{
		ID:        blob.ID,
		Data:      base64.StdEncoding.EncodeToString(blob.Data),
		Size:      fmt.Sprintf("%d", metadata.Size),
		CreatedAt: metadata.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}
