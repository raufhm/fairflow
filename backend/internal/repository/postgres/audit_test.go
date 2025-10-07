package postgres_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/internal/repository/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestAuditLogRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	auditRepo := postgres.NewAuditLogRepository(bunDB)

	userID := int64(1)
	log := &domain.AuditLog{
		UserID:   &userID,
		Action:   "test action",
		UserName: "test user",
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`INSERT INTO "audit_logs"`).WillReturnRows(rows)

	err = auditRepo.Create(log)

	assert.NoError(t, err)
}

func TestAuditLogRepository_GetRecent(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	auditRepo := postgres.NewAuditLogRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "audit_logs"`).WillReturnRows(rows)

	_, err = auditRepo.GetRecent(1)

	assert.NoError(t, err)
}
