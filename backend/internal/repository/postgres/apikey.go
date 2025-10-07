package postgres

import (
	"context"
	"time"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/uptrace/bun"
)

type apiKeyRepository struct {
	db *bun.DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *bun.DB) domain.APIKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) Create(apiKey *domain.APIKey) error {
	ctx := context.Background()
	apiKey.CreatedAt = time.Now()
	_, err := r.db.NewInsert().Model(apiKey).Exec(ctx)
	return err
}

func (r *apiKeyRepository) GetByID(id int64) (*domain.APIKey, error) {
	ctx := context.Background()
	apiKey := &domain.APIKey{Active: true}
	err := r.db.NewSelect().Model(apiKey).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (r *apiKeyRepository) GetByHash(hash string) (*domain.APIKey, error) {
	ctx := context.Background()
	apiKey := &domain.APIKey{Active: true}
	err := r.db.NewSelect().Model(apiKey).Where("key_hash = ?", hash).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (r *apiKeyRepository) GetByUserID(userID int64) ([]*domain.APIKey, error) {
	ctx := context.Background()
	var apiKeys []*domain.APIKey
	err := r.db.NewSelect().
		Model(&apiKeys).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	// Set Active field for all keys
	for _, key := range apiKeys {
		key.Active = true
	}

	return apiKeys, nil
}

func (r *apiKeyRepository) Delete(id int64) error {
	ctx := context.Background()
	_, err := r.db.NewDelete().Model(&domain.APIKey{}).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *apiKeyRepository) UpdateLastUsed(id int64) error {
	ctx := context.Background()
	_, err := r.db.NewUpdate().
		Model(&domain.APIKey{}).
		Set("last_used_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)
	return err
}
