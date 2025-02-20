package postgres

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {

	mockDB, mock, err := sqlmock.New()
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	assert.NoError(t, err)
	assert.NotNil(t, mock)
	db := &Postgres{
		Db: sqlxDB,
	}
	assert.NotEmpty(t, db)
}
