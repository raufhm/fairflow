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

func TestAssignmentRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	assignmentRepo := postgres.NewAssignmentRepository(bunDB)

	assignment := &domain.Assignment{
		GroupID:  1,
		MemberID: 1,
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`INSERT INTO "assignments"`).WillReturnRows(rows)

	err = assignmentRepo.Create(assignment)

	assert.NoError(t, err)
}

func TestAssignmentRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	assignmentRepo := postgres.NewAssignmentRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT (.+) FROM "assignments"`).WillReturnRows(rows)

	_, err = assignmentRepo.GetByID(1)

	assert.NoError(t, err)
}

func TestAssignmentRepository_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	assignmentRepo := postgres.NewAssignmentRepository(bunDB)

	mock.ExpectExec(`UPDATE "assignments"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = assignmentRepo.UpdateStatus(1, domain.AssignmentStatusCompleted)

	assert.NoError(t, err)
}

func TestAssignmentRepository_GetByGroupID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	assignmentRepo := postgres.NewAssignmentRepository(bunDB)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT a.id, a.metadata, a.created_at, m.id as member_id, m.name as member_name FROM assignments AS a JOIN members AS m ON a.member_id = m.id WHERE (.+)`).WillReturnRows(rows)

	_, err = assignmentRepo.GetByGroupID(1, 10, 0)

	assert.NoError(t, err)
}

func TestAssignmentRepository_GetCountByGroupID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	assignmentRepo := postgres.NewAssignmentRepository(bunDB)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT count(.+) FROM "assignments"`).WillReturnRows(rows)

	_, err = assignmentRepo.GetCountByGroupID(1)

	assert.NoError(t, err)
}

func TestAssignmentRepository_GetCountsByMemberIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	bunDB := bun.NewDB(db, pgdialect.New())
	assignmentRepo := postgres.NewAssignmentRepository(bunDB)

	rows := sqlmock.NewRows([]string{"member_id", "count"}).AddRow(1, 1)
	mock.ExpectQuery(`SELECT member_id, COUNT(.+) as count FROM assignments WHERE (.+) GROUP BY "member_id"`).WillReturnRows(rows)

	_, err = assignmentRepo.GetCountsByMemberIDs([]int64{1})

	assert.NoError(t, err)
}
