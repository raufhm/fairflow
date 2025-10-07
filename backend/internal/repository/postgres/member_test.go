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

func TestMemberRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	memberRepo := postgres.NewMemberRepository(bunDB)

	member := &domain.Member{
		GroupID: 1,
		Name:    "Test Member",
		Weight:  100,
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`INSERT INTO "members"`).WillReturnRows(rows)

	err = memberRepo.Create(member)

	assert.NoError(t, err)
}

func TestMemberRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	memberRepo := postgres.NewMemberRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "members"`).WillReturnRows(rows)

	_, err = memberRepo.GetByID(1)

	assert.NoError(t, err)
}

func TestMemberRepository_GetByGroupID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	memberRepo := postgres.NewMemberRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "members"`).WillReturnRows(rows)

	_, err = memberRepo.GetByGroupID(1)

	assert.NoError(t, err)
}

func TestMemberRepository_GetActiveByGroupID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	memberRepo := postgres.NewMemberRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "members"`).WillReturnRows(rows)

	_, err = memberRepo.GetActiveByGroupID(1)

	assert.NoError(t, err)
}

func TestMemberRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	memberRepo := postgres.NewMemberRepository(bunDB)

	member := &domain.Member{
		ID:      1,
		GroupID: 1,
		Name:    "Test Member",
		Weight:  100,
	}

	mock.ExpectExec(`UPDATE "members"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = memberRepo.Update(member)

	assert.NoError(t, err)
}

func TestMemberRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	memberRepo := postgres.NewMemberRepository(bunDB)

	mock.ExpectExec(`DELETE FROM "members"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = memberRepo.Delete(1)

	assert.NoError(t, err)
}

func TestMemberRepository_IncrementOpenAssignments(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	memberRepo := postgres.NewMemberRepository(bunDB)

	mock.ExpectExec(`UPDATE "members"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = memberRepo.IncrementOpenAssignments(1)

	assert.NoError(t, err)
}

func TestMemberRepository_DecrementOpenAssignments(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	memberRepo := postgres.NewMemberRepository(bunDB)

	mock.ExpectExec(`UPDATE "members"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = memberRepo.DecrementOpenAssignments(1)

	assert.NoError(t, err)
}

func TestMemberRepository_GetDailyAssignmentCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	memberRepo := postgres.NewMemberRepository(bunDB)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT count(.+) FROM "assignments"`).WillReturnRows(rows)

	_, err = memberRepo.GetDailyAssignmentCount(1)

	assert.NoError(t, err)
}
