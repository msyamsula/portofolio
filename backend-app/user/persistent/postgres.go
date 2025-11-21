package persistent

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

type postgres struct {
	*sqlx.DB
}

func (s *postgres) InsertUser(c context.Context, user User) (User, error) {
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

	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(ctx, QueryInsertUser)
	if err != nil {
		return user, err
	}

	var rows *sql.Rows
	rows, err = stmt.QueryContext(ctx, map[string]interface{}{
		"username": user.Username,
		"online":   user.Online,
	})
	if err != nil {
		return User{}, err
	}

	if rows.Next() {
		err = rows.Scan(&user.Id)
		if err != nil {
			return User{}, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return user, err
	}
	return user, nil
}

func (s *postgres) GetUser(c context.Context, username string) (User, error) {
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
	err = row.Scan(&user.Id, &user.Username, &user.Online)
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
