package metadata

import (
	"database/sql"
	"fmt"
	"time"

	"simple-drive/internal/models"

	_ "modernc.org/sqlite"
)

type Service struct {
	db *sql.DB
}

func NewService(databaseURL string) (*Service, error) {
	db, err := sql.Open("sqlite", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata database: %w", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS blob_metadata (
		id TEXT PRIMARY KEY,
		size INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		storage_type TEXT NOT NULL,
		storage_path TEXT NOT NULL
	);
	`
	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("failed to create metadata table: %w", err)
	}

	return &Service{db: db}, nil
}

func (s *Service) Save(metadata *models.Metadata) error {
	query := `INSERT OR REPLACE INTO blob_metadata (id, size, created_at, storage_type, storage_path) 
	          VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query,
		metadata.ID,
		metadata.Size,
		metadata.CreatedAt.Format(time.RFC3339),
		metadata.StorageType,
		metadata.StoragePath,
	)
	if err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}
	return nil
}

func (s *Service) Get(id string) (*models.Metadata, error) {
	query := `SELECT id, size, created_at, storage_type, storage_path 
	          FROM blob_metadata WHERE id = ?`

	var metadata models.Metadata
	var createdAtStr string

	err := s.db.QueryRow(query, id).Scan(
		&metadata.ID,
		&metadata.Size,
		&createdAtStr,
		&metadata.StorageType,
		&metadata.StoragePath,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("metadata not found: %s", id)
		}
		return nil, fmt.Errorf("failed to retrieve metadata: %w", err)
	}

	metadata.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		metadata.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}
	}

	return &metadata, nil
}

func (s *Service) Close() error {
	return s.db.Close()
}
