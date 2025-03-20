package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/msyamsula/portofolio/tech-stack/postgres"
	"go.opentelemetry.io/otel"
)

type Persistence struct {
	*postgres.Postgres
}

type User struct {
	Username string `json:"username,omitempty"`
	Id       int64  `json:"id,omitempty"`
}

func (s *Persistence) InsertUser(c context.Context, username string) (User, error) {
	ctx, span := otel.Tracer("").Start(c, "repository.persistence.InsertUser")
	defer span.End()

	tx := s.MustBeginTx(ctx, nil)
	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
			tx.Rollback()
		}
	}()

	var user User
	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(ctx, QueryInsertUser)
	if err != nil {
		return user, err
	}

	var row *sqlx.Row
	row = stmt.QueryRowContext(ctx, map[string]interface{}{
		"username": username,
	})

	user.Username = username
	err = row.Scan(&user.Id)
	if err != nil {
		return User{}, err
	}

	err = tx.Commit()
	if err != nil {
		return user, err
	}
	return user, nil
}

func (s *Persistence) GetUser(c context.Context, username string) (User, error) {
	ctx, span := otel.Tracer("").Start(c, "repository.persistence.GetUser")
	defer span.End()

	var err error
	tx := s.MustBeginTx(ctx, nil)
	defer func() {
		if err != nil {
			span.RecordError(err)
			tx.Rollback()
		}
	}()

	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(ctx, QueryGetUser)
	if err != nil {
		return User{}, err
	}

	var row *sqlx.Row
	row = stmt.QueryRowContext(ctx, map[string]interface{}{
		"username": username,
	})

	var user User
	err = row.Scan(&user.Id, &user.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrUserNotFound
		}
		return user, err
	}

	err = tx.Commit()
	if err != nil {
		return User{}, err
	}
	return user, nil
}
