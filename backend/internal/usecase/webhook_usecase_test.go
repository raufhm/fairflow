package usecase_test

import (
	"testing"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func TestCreateWebhook(t *testing.T) {
	webhookRepo := newMockWebhookRepo()
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewWebhookUseCase(webhookRepo, auditRepo)

	webhook, err := uc.CreateWebhook(1, 1, "http://example.com", []string{"assignment.created"})

	assert.NoError(t, err)
	assert.NotNil(t, webhook)
	assert.Equal(t, "http://example.com", webhook.URL)
}

func TestGetWebhooksByGroup(t *testing.T) {
	webhookRepo := newMockWebhookRepo()
	webhookRepo.webhooks[1] = &domain.Webhook{ID: 1, GroupID: 1, URL: "http://example.com"}
	webhookRepo.webhooks[2] = &domain.Webhook{ID: 2, GroupID: 1, URL: "http://example2.com"}
	uc := usecase.NewWebhookUseCase(webhookRepo, nil)

	webhooks, err := uc.GetWebhooksByGroup(1)

	assert.NoError(t, err)
	assert.Len(t, webhooks, 2)
}

func TestUpdateWebhook(t *testing.T) {
	webhookRepo := newMockWebhookRepo()
	webhookRepo.webhooks[1] = &domain.Webhook{ID: 1, GroupID: 1, URL: "http://example.com"}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewWebhookUseCase(webhookRepo, auditRepo)

	updatedWebhook := &domain.Webhook{ID: 1, GroupID: 1, URL: "http://new-example.com"}
	err := uc.UpdateWebhook(1, updatedWebhook)

	assert.NoError(t, err)
	webhook, _ := webhookRepo.webhooks[1]
	assert.Equal(t, "http://new-example.com", webhook.URL)
}

func TestDeleteWebhook(t *testing.T) {
	webhookRepo := newMockWebhookRepo()
	webhookRepo.webhooks[1] = &domain.Webhook{ID: 1, GroupID: 1, URL: "http://example.com"}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewWebhookUseCase(webhookRepo, auditRepo)

	err := uc.DeleteWebhook(1, 1)

	assert.NoError(t, err)
	_, ok := webhookRepo.webhooks[1]
	assert.False(t, ok)
}
