package local

import (
	"fmt"
	"os"
	"path/filepath"

	"simple-drive/internal/models"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &LocalStorage{basePath: basePath}, nil
}

func (ls *LocalStorage) Save(blob *models.Blob) (string, error) {
	filePath := filepath.Join(ls.basePath, blob.ID)
	if err := os.WriteFile(filePath, blob.Data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return filePath, nil
}

func (ls *LocalStorage) Retrieve(id string) (*models.Blob, error) {
	filePath := filepath.Join(ls.basePath, id)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return &models.Blob{
		ID:   id,
		Data: data,
	}, nil
}
