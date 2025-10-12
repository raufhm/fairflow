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

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	userRepo := postgres.NewUserRepository(bunDB)

	user := &domain.User{
		Name:         "Test User",
		Email:        "test31@example.com",
		PasswordHash: "hash",
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`INSERT INTO "users"`).WillReturnRows(rows)

	err = userRepo.Create(context.Background(), user)

	assert.NoError(t, err)
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	userRepo := postgres.NewUserRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "users"`).WillReturnRows(rows)

	_, err = userRepo.GetByID(context.Background(), 1)

	assert.NoError(t, err)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	userRepo := postgres.NewUserRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "users"`).WillReturnRows(rows)

	_, err = userRepo.GetByEmail(context.Background(), "test@example.com")

	assert.NoError(t, err)
}

func TestUserRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	userRepo := postgres.NewUserRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "users"`).WillReturnRows(rows)

	_, err = userRepo.GetAll(context.Background())

	assert.NoError(t, err)
}

func TestUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	userRepo := postgres.NewUserRepository(bunDB)

	user := &domain.User{
		ID:           1,
		Name:         "Test User",
		Email:        "test36@example.com",
		PasswordHash: "hash",
	}

	mock.ExpectExec(`UPDATE "users"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = userRepo.Update(context.Background(), user)

	assert.NoError(t, err)
}

func TestUserRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	userRepo := postgres.NewUserRepository(bunDB)

	mock.ExpectExec(`DELETE FROM "users"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = userRepo.Delete(context.Background(), 1)

	assert.NoError(t, err)
}

func TestUserRepository_UpdateRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	userRepo := postgres.NewUserRepository(bunDB)

	mock.ExpectExec(`UPDATE "users"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = userRepo.UpdateRole(context.Background(), 1, domain.RoleAdmin)

	assert.NoError(t, err)
}
