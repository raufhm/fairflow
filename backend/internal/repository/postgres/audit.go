package postgres

import (
	"context"
	"time"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/uptrace/bun"
)

type auditLogRepository struct {
	db *bun.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *bun.DB) domain.AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(log *domain.AuditLog) error {
	ctx := context.Background()
	log.CreatedAt = time.Now()
	_, err := r.db.NewInsert().Model(log).Exec(ctx)
	return err
}

func (r *auditLogRepository) GetRecent(limit int) ([]*domain.AuditLog, error) {
	ctx := context.Background()
	var logs []*domain.AuditLog
	err := r.db.NewSelect().
		Model(&logs).
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)
	return logs, err
}
