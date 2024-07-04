package storage

import (
	"fmt"

	"github.com/arhsxro/platform-go-challenge/config"
	"github.com/arhsxro/platform-go-challenge/models"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sqlx.DB
}

func NewPostgresStore(cfg *config.Config) (*PostgresStore, error) {
	db, err := NewPostgresDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return &PostgresStore{db: db}, nil
}

// Retrieves a user's favorite assets from the database
func (store *PostgresStore) GetUserFavorites(userID string) ([]models.Asset, error) {
	var assets []models.Asset
	query := "SELECT asset_id, type, description, data FROM assets WHERE user_id = $1"
	err := store.db.Select(&assets, query, userID)
	return assets, err
}

// Adds a new favorite asset to a user in the database
func (store *PostgresStore) AddFavorite(userID string, asset models.Asset) error {
	tx, err := store.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
        INSERT INTO assets (user_id, asset_id, type, description, data)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (asset_id) DO NOTHING`

	_, err = tx.Exec(query, userID, asset.ID, asset.Type, asset.Description, string(asset.Data))
	if err != nil {
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Removes an asset from a user in the database
func (store *PostgresStore) RemoveFavorite(userID, assetID string) error {
	query := "DELETE FROM assets WHERE user_id = $1 AND asset_id = $2"
	_, err := store.db.Exec(query, userID, assetID)
	return err
}

// Updates an asset's description from a user in the database
func (store *PostgresStore) UpdateFavoriteDescription(userID, assetID, newDescription string) error {
	query := "UPDATE assets SET description = $1 WHERE user_id = $2 AND asset_id = $3"
	_, err := store.db.Exec(query, newDescription, userID, assetID)
	return err
}
