package test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type RepositoryTestSuite struct {
	suite.Suite
	suite.SetupAllSuite
	suite.TearDownAllSuite

	mockErr error
	ctrl    *gomock.Controller

	mock   sqlmock.Sqlmock
	sqlxDb *sqlx.DB
	mockDb *sql.DB
}

func (s *RepositoryTestSuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())
	s.mockErr = errors.New("ops")

	mockDb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	s.Nil(err)

	sqlxDb := sqlx.NewDb(mockDb, "sqlmock")
	s.mock = mock
	s.sqlxDb = sqlxDb
	s.mockDb = mockDb
}
func (s *RepositoryTestSuite) TearDownSuite() {
	s.ctrl.Finish()
	s.mockDb.Close()
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
