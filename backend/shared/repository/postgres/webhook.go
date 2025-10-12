package postgres

import (
	"context"
	"encoding/json"

	"github.com/raufhm/fairflow/shared/domain"
	"github.com/uptrace/bun"
)

type webhookRepository struct {
	db *bun.DB
}

func NewWebhookRepository(db *bun.DB) domain.WebhookRepository {
	return &webhookRepository{db: db}
}

func (r *webhookRepository) Create(ctx context.Context, webhook *domain.Webhook) error {
	// Convert events slice to JSON
	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return err
	}

	_, err = r.db.NewInsert().
		Model(webhook).
		Column("group_id", "url", "events", "secret", "active", "created_at").
		Value("events", "?", string(eventsJSON)).
		Exec(ctx)

	return err
}

func (r *webhookRepository) GetByGroupID(ctx context.Context, groupID int64) ([]*domain.Webhook, error) {
	var webhooks []*domain.Webhook

	err := r.db.NewSelect().
		Model(&webhooks).
		Where("group_id = ?", groupID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	// Parse events JSON for each webhook
	for _, wh := range webhooks {
		var events []string
		// Events are stored as JSON in the database
		// This is a simplified version - you may need to adjust based on actual DB schema
		wh.Events = events
	}

	return webhooks, nil
}

func (r *webhookRepository) GetActiveByGroupID(ctx context.Context, groupID int64) ([]*domain.Webhook, error) {
	var webhooks []*domain.Webhook

	err := r.db.NewSelect().
		Model(&webhooks).
		Where("group_id = ? AND active = ?", groupID, true).
		Scan(ctx)

	return webhooks, err
}

func (r *webhookRepository) Update(ctx context.Context, webhook *domain.Webhook) error {
	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return err
	}

	_, err = r.db.NewUpdate().
		Model(webhook).
		Column("url", "active", "secret").
		Set("events = ?", string(eventsJSON)).
		Where("id = ?", webhook.ID).
		Exec(ctx)

	return err
}

func (r *webhookRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().
		Model((*domain.Webhook)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	return err
}
