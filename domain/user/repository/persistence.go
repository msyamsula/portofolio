package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/msyamsula/portofolio/tech-stack/postgres"
)

type Persistence struct {
	*postgres.Postgres
}

type User struct {
	Username string `json:"username,omitempty"`
	Id       int64  `json:"id,omitempty"`
}

func (s *Persistence) InsertUser(c context.Context, username string) (User, error) {
	var err error
	tx := s.MustBeginTx(c, nil)
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	var user User
	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(c, QueryInsertUser)
	if err != nil {
		return user, err
	}

	var row *sqlx.Row
	row = stmt.QueryRowContext(c, map[string]interface{}{
		"username": username,
	})
	if row == nil {
		return user, ErrUserNotFound
	}

	user.Username = username
	err = row.Scan(&user.Id)
	if err != nil {
		return User{}, err
	}

	err = tx.Commit()
	return user, err
}

func (s *Persistence) GetUser(c context.Context, username string) (User, error) {
	var err error
	tx := s.MustBeginTx(c, nil)
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(c, QueryGetUser)
	if err != nil {
		return User{}, err
	}

	var rows *sql.Rows
	rows, err = stmt.QueryContext(c, map[string]interface{}{
		"username": username,
	})

	if rows == nil {
		return User{}, ErrUserNotFound
	}

	var user User
	count := 0
	for rows.Next() {
		err = rows.Scan(&user.Id, &user.Username)
		if err != nil {
			return User{}, err
		}
		count++
		break
	}

	if count == 0 {
		return User{}, ErrUserNotFound
	}

	err = tx.Commit()
	return user, err
}
