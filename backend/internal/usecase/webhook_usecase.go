package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/raufhm/rra/internal/domain"
)

type WebhookUseCase struct {
	webhookRepo domain.WebhookRepository
	auditRepo   domain.AuditLogRepository
}

func NewWebhookUseCase(webhookRepo domain.WebhookRepository, auditRepo domain.AuditLogRepository) *WebhookUseCase {
	return &WebhookUseCase{
		webhookRepo: webhookRepo,
		auditRepo:   auditRepo,
	}
}

// CreateWebhook creates a new webhook for a group
func (uc *WebhookUseCase) CreateWebhook(userID, groupID int64, url string, events []string) (*domain.Webhook, error) {
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

	if err := uc.webhookRepo.Create(webhook); err != nil {
		return nil, err
	}

	// Audit log
	resourceType := "webhook"
	details := fmt.Sprintf("Created webhook for group %d", groupID)
	uc.auditRepo.Create(&domain.AuditLog{
		UserID:       &userID,
		Action:       "webhook_created",
		ResourceType: &resourceType,
		ResourceID:   &webhook.ID,
		Details:      &details,
		CreatedAt:    time.Now(),
	})

	return webhook, nil
}

// GetWebhooksByGroup returns all webhooks for a group
func (uc *WebhookUseCase) GetWebhooksByGroup(groupID int64) ([]*domain.Webhook, error) {
	return uc.webhookRepo.GetByGroupID(groupID)
}

// UpdateWebhook updates a webhook
func (uc *WebhookUseCase) UpdateWebhook(userID int64, webhook *domain.Webhook) error {
	if err := uc.webhookRepo.Update(webhook); err != nil {
		return err
	}

	// Audit log
	resourceType := "webhook"
	details := fmt.Sprintf("Updated webhook %d", webhook.ID)
	uc.auditRepo.Create(&domain.AuditLog{
		UserID:       &userID,
		Action:       "webhook_updated",
		ResourceType: &resourceType,
		ResourceID:   &webhook.ID,
		Details:      &details,
		CreatedAt:    time.Now(),
	})

	return nil
}

// DeleteWebhook deletes a webhook
func (uc *WebhookUseCase) DeleteWebhook(userID, webhookID int64) error {
	if err := uc.webhookRepo.Delete(webhookID); err != nil {
		return err
	}

	// Audit log
	resourceType := "webhook"
	details := fmt.Sprintf("Deleted webhook %d", webhookID)
	uc.auditRepo.Create(&domain.AuditLog{
		UserID:       &userID,
		Action:       "webhook_deleted",
		ResourceType: &resourceType,
		ResourceID:   &webhookID,
		Details:      &details,
		CreatedAt:    time.Now(),
	})

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
