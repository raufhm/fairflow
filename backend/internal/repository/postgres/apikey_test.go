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

func TestAPIKeyRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	apiKeyRepo := postgres.NewAPIKeyRepository(bunDB)

	apiKey := &domain.APIKey{
		UserID:  1,
		Name:    "Test Key",
		KeyHash: "hash",
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`INSERT INTO "api_keys"`).WillReturnRows(rows)

	err = apiKeyRepo.Create(apiKey)

	assert.NoError(t, err)
}

func TestAPIKeyRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	apiKeyRepo := postgres.NewAPIKeyRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "api_keys"`).WillReturnRows(rows)

	_, err = apiKeyRepo.GetByID(1)

	assert.NoError(t, err)
}

func TestAPIKeyRepository_GetByHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	apiKeyRepo := postgres.NewAPIKeyRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "api_keys"`).WillReturnRows(rows)

	_, err = apiKeyRepo.GetByHash("hash")

	assert.NoError(t, err)
}

func TestAPIKeyRepository_GetByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	apiKeyRepo := postgres.NewAPIKeyRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "api_keys"`).WillReturnRows(rows)

	_, err = apiKeyRepo.GetByUserID(1)

	assert.NoError(t, err)
}

func TestAPIKeyRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	apiKeyRepo := postgres.NewAPIKeyRepository(bunDB)

	mock.ExpectExec(`DELETE FROM "api_keys"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = apiKeyRepo.Delete(1)

	assert.NoError(t, err)
}

func TestAPIKeyRepository_UpdateLastUsed(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	apiKeyRepo := postgres.NewAPIKeyRepository(bunDB)

	mock.ExpectExec(`UPDATE "api_keys"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = apiKeyRepo.UpdateLastUsed(1)

	assert.NoError(t, err)
}
