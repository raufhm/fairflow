package postgres_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raufhm/fairflow/shared/domain"
	"github.com/raufhm/fairflow/shared/repository/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestGroupRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	groupRepo := postgres.NewGroupRepository(bunDB)

	group := &domain.Group{
		UserID:   1,
		Name:     "Test Group",
		Strategy: domain.StrategyWeightedRoundRobin,
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`INSERT INTO "groups"`).WillReturnRows(rows)

	err = groupRepo.Create(context.Background(), group)

	assert.NoError(t, err)
}

func TestGroupRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	groupRepo := postgres.NewGroupRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "groups"`).WillReturnRows(rows)

	_, err = groupRepo.GetByID(context.Background(), 1)

	assert.NoError(t, err)
}

func TestGroupRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	groupRepo := postgres.NewGroupRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "groups"`).WillReturnRows(rows)

	_, err = groupRepo.GetAll(context.Background())

	assert.NoError(t, err)
}

func TestGroupRepository_GetByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	groupRepo := postgres.NewGroupRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "groups"`).WillReturnRows(rows)

	_, err = groupRepo.GetByUserID(context.Background(), 1)

	assert.NoError(t, err)
}

func TestGroupRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	groupRepo := postgres.NewGroupRepository(bunDB)

	group := &domain.Group{
		ID:       1,
		UserID:   1,
		Name:     "Test Group",
		Strategy: domain.StrategyWeightedRoundRobin,
	}

	mock.ExpectExec(`UPDATE "groups"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = groupRepo.Update(context.Background(), group)

	assert.NoError(t, err)
}

func TestGroupRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	groupRepo := postgres.NewGroupRepository(bunDB)

	mock.ExpectExec(`DELETE FROM "groups"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = groupRepo.Delete(context.Background(), 1)

	assert.NoError(t, err)
}
