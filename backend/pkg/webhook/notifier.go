package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/raufhm/fairflow/internal/domain"
)

// Notifier handles webhook notifications
type Notifier struct {
	webhookRepo domain.WebhookRepository
	httpClient  *http.Client
}

// NewNotifier creates a new webhook notifier
func NewNotifier(webhookRepo domain.WebhookRepository) *Notifier {
	return &Notifier{
		webhookRepo: webhookRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Event represents a webhook event
type Event struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Notify sends a webhook notification to all registered webhooks for a group
func (n *Notifier) Notify(ctx context.Context, groupID int64, eventType string, data map[string]interface{}) error {
	webhooks, err := n.webhookRepo.GetActiveByGroupID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("failed to get webhooks: %w", err)
	}

	event := Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	for _, webhook := range webhooks {
		// Check if webhook is subscribed to this event type
		if !contains(webhook.Events, eventType) {
			continue
		}

		go n.sendWebhook(ctx, webhook, event)
	}

	return nil
}

// sendWebhook sends a single webhook request
func (n *Notifier) sendWebhook(ctx context.Context, webhook *domain.Webhook, event Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Failed to marshal webhook payload: %v\n", err)
		return
	}

	// Create HMAC signature
	signature := generateSignature(payload, webhook.Secret)

	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Failed to create webhook request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", event.Type)

	resp, err := n.httpClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to send webhook to %s: %v\n", webhook.URL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		fmt.Printf("Webhook %s returned error status: %d\n", webhook.URL, resp.StatusCode)
	}
}

// generateSignature creates an HMAC signature for webhook validation
func generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
