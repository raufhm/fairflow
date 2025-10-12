package messaging

import "time"

// AssignmentCreatedEvent is published when a new assignment is created
type AssignmentCreatedEvent struct {
	AssignmentID int64     `json:"assignment_id"`
	GroupID      int64     `json:"group_id"`
	MemberID     int64     `json:"member_id"`
	Metadata     *string   `json:"metadata,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuditLogEvent is published for audit logging
type AuditLogEvent struct {
	UserID       *int64    `json:"user_id,omitempty"`
	UserName     string    `json:"user_name"`
	Action       string    `json:"action"`
	ResourceType *string   `json:"resource_type,omitempty"`
	ResourceID   *int64    `json:"resource_id,omitempty"`
	Details      *string   `json:"details,omitempty"`
	IPAddress    string    `json:"ip_address"`
	CreatedAt    time.Time `json:"created_at"`
}

// WebhookTriggerEvent is published when a webhook needs to be triggered
type WebhookTriggerEvent struct {
	WebhookID    int64       `json:"webhook_id"`
	EventType    string      `json:"event_type"`
	Payload      interface{} `json:"payload"`
	TriggeredAt  time.Time   `json:"triggered_at"`
}
