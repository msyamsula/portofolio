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
	Username string `json:"username"`
	Id       int64  `json:"id"`
	Online   bool   `json:"online"`
	Unread   int64  `json:"unread"`
}

func (s *Persistence) InsertUser(c context.Context, user User) (User, error) {
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

	for rows.Next() {
		err = rows.Scan(&user.Id)
		if err != nil {
			return User{}, err
		}
		break
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

func (s *Persistence) AddFriend(c context.Context, userA, userB User) error {

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

func (s *Persistence) GetFriends(c context.Context, user User) ([]User, error) {

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
