package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/raufhm/fairflow/shared/domain"
)

type WebhookUseCase struct {
	webhookRepo domain.WebhookRepository
}

func NewWebhookUseCase(webhookRepo domain.WebhookRepository) *WebhookUseCase {
	return &WebhookUseCase{
		webhookRepo: webhookRepo,
	}
}

// CreateWebhook creates a new webhook for a group
func (uc *WebhookUseCase) CreateWebhook(ctx context.Context, userID, groupID int64, userName, url string, events []string) (*domain.Webhook, error) {
	// Generate secret for webhook validation
	secret, err := generateWebhookSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate webhook secret: %w", err)
	}

	webhook := &domain.Webhook{
		GroupID:   groupID,
		URL:       url,
		Events:    events,
		Secret:    secret,
		Active:    true,
		CreatedAt: time.Now(),
	}

	if err := uc.webhookRepo.Create(ctx, webhook); err != nil {
		return nil, err
	}

	return webhook, nil
}

// GetWebhooksByGroup returns all webhooks for a group
func (uc *WebhookUseCase) GetWebhooksByGroup(ctx context.Context, groupID int64) ([]*domain.Webhook, error) {
	return uc.webhookRepo.GetByGroupID(ctx, groupID)
}

// UpdateWebhook updates a webhook
func (uc *WebhookUseCase) UpdateWebhook(ctx context.Context, userID int64, userName string, webhook *domain.Webhook) error {
	if err := uc.webhookRepo.Update(ctx, webhook); err != nil {
		return err
	}

	return nil
}

// DeleteWebhook deletes a webhook
func (uc *WebhookUseCase) DeleteWebhook(ctx context.Context, userID int64, userName string, webhookID int64) error {
	if err := uc.webhookRepo.Delete(ctx, webhookID); err != nil {
		return err
	}

	return nil
}

// generateWebhookSecret generates a random secret for webhook validation
func generateWebhookSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
