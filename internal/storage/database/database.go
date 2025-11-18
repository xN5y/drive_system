package database

import (
	"database/sql"
	"fmt"

	"simple-drive/internal/models"

	_ "modernc.org/sqlite"
)

type DatabaseStorage struct {
	db *sql.DB
}

func NewDatabaseStorage(databaseURL string) (*DatabaseStorage, error) {
	db, err := sql.Open("sqlite", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS blob_storage (
		id TEXT PRIMARY KEY,
		data BLOB NOT NULL
	);
	`
	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &DatabaseStorage{db: db}, nil
}

func (ds *DatabaseStorage) Save(blob *models.Blob) (string, error) {
	query := `INSERT OR REPLACE INTO blob_storage (id, data) VALUES (?, ?)`
	if _, err := ds.db.Exec(query, blob.ID, blob.Data); err != nil {
		return "", fmt.Errorf("failed to save blob: %w", err)
	}
	return blob.ID, nil
}

func (ds *DatabaseStorage) Retrieve(id string) (*models.Blob, error) {
	query := `SELECT id, data FROM blob_storage WHERE id = ?`
	var blob models.Blob
	err := ds.db.QueryRow(query, id).Scan(&blob.ID, &blob.Data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blob not found: %s", id)
		}
		return nil, fmt.Errorf("failed to retrieve blob: %w", err)
	}
	return &blob, nil
}

func (ds *DatabaseStorage) Close() error {
	return ds.db.Close()
}
