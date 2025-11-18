package storage

import "simple-drive/internal/models"

type Storage interface {
	Save(blob *models.Blob) (string, error)
	Retrieve(id string) (*models.Blob, error)
}
