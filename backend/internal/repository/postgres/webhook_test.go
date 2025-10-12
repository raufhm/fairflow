package postgres_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/internal/repository/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestWebhookRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	webhookRepo := postgres.NewWebhookRepository(bunDB)

	webhook := &domain.Webhook{
		GroupID: 1,
		URL:     "http://example.com",
		Events:  []string{"assignment.created"},
		Secret:  "secret",
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`INSERT INTO "webhooks"`).WillReturnRows(rows)

	err = webhookRepo.Create(context.Background(), webhook)

	assert.NoError(t, err)
}

func TestWebhookRepository_GetByGroupID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	webhookRepo := postgres.NewWebhookRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "webhooks"`).WillReturnRows(rows)

	_, err = webhookRepo.GetByGroupID(context.Background(), 1)

	assert.NoError(t, err)
}

func TestWebhookRepository_GetActiveByGroupID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	webhookRepo := postgres.NewWebhookRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "webhooks"`).WillReturnRows(rows)

	_, err = webhookRepo.GetActiveByGroupID(context.Background(), 1)

	assert.NoError(t, err)
}

func TestWebhookRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	webhookRepo := postgres.NewWebhookRepository(bunDB)

	webhook := &domain.Webhook{
		ID:      1,
		GroupID: 1,
		URL:     "http://example.com",
		Events:  []string{"assignment.created"},
		Secret:  "secret",
	}

	mock.ExpectExec(`UPDATE "webhooks"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = webhookRepo.Update(context.Background(), webhook)

	assert.NoError(t, err)
}

func TestWebhookRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	webhookRepo := postgres.NewWebhookRepository(bunDB)

	mock.ExpectExec(`DELETE FROM "webhooks"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = webhookRepo.Delete(context.Background(), 1)

	assert.NoError(t, err)
}
