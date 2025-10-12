package postgres

import (
	"context"
	"time"

	"github.com/raufhm/fairflow/shared/domain"
	"github.com/uptrace/bun"
)

type auditLogRepository struct {
	db *bun.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *bun.DB) domain.AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	log.CreatedAt = time.Now()
	_, err := r.db.NewInsert().Model(log).Exec(ctx)
	return err
}

func (r *auditLogRepository) GetRecent(ctx context.Context, limit int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	err := r.db.NewSelect().
		Model(&logs).
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)
	return logs, err
}
