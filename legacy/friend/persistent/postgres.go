package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

type postgres struct {
	*sqlx.DB
}

func (s *postgres) AddFriend(c context.Context, userA, userB User) error {

	ctx, span := otel.Tracer("").Start(c, "repository.persistence.AddFriend")
	defer span.End()

	tx := s.MustBeginTx(ctx, nil)
	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
			tx.Rollback()
		}
	}()

	if userA.Id == userB.Id {
		err = ErrIdMustBeDifferent
		return err
	}

	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(ctx, QueryAddFriend)
	if err != nil {
		return err
	}

	var rows *sql.Rows
	rows, err = stmt.QueryContext(ctx, map[string]interface{}{
		"small_id": min(userA.Id, userB.Id),
		"big_id":   max(userA.Id, userB.Id),
	})
	if err != nil {
		return err
	}

	var id int64
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (s *postgres) GetFriends(c context.Context, user User) ([]User, error) {

	ctx, span := otel.Tracer("").Start(c, "repository.persistence.GetFriends")
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
	stmt, err = tx.PrepareNamedContext(ctx, QueryGetFriends)
	if err != nil {
		return []User{}, err
	}

	var rows *sql.Rows
	rows, err = stmt.QueryContext(ctx, map[string]interface{}{
		"user_id": user.Id,
	})
	if err != nil {
		return []User{}, err
	}

	users := []User{}
	for rows.Next() {
		tmp := User{}
		err = rows.Scan(&tmp.Id, &tmp.Username, &tmp.Online, &tmp.Unread)
		if err != nil {
			continue
		}
		users = append(users, tmp)
	}

	err = tx.Commit()
	if err != nil {
		return []User{}, err
	}
	return users, nil
}
