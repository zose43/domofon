package postgres

import (
	"context"
	"database/sql"
	"domofon/internal/domain/models"
	"domofon/internal/storage"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db *sql.DB
}

const (
	UniqueViolationErr = "23505"
)

func NewStorage(dsn string) (*Storage, error) {
	const op = "storage.postgres.new"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.postgres.saveUser"

	stmt, err := s.db.Prepare("insert into users (email, pass_hash) VALUES (?,?)")
	if err != nil {
		return 0, fmt.Errorf("%s %w", op, err)
	}

	result, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		var postgresErr *pgconn.PgError
		if errors.As(err, &postgresErr) && postgresErr.Code == UniqueViolationErr {
			return 0, fmt.Errorf("%s %w", op, storage.ErrUserExists)
		}

		return 0, fmt.Errorf("%s %w", op, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s %w", op, err)
	}
	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.user"

	stmt, err := s.db.Prepare("select * from users where email = ?")
	if err != nil {
		return models.User{}, fmt.Errorf("%s %w", op, err)
	}

	result := stmt.QueryRowContext(ctx, email)

	var user models.User
	err = result.Scan(&user.Id, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s %w", op, storage.ErrNotFound)
		}

		return models.User{}, fmt.Errorf("%s %w", op, err)
	}

	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.postgres.isAdmin"

	stmt, err := s.db.Prepare("select users.is_admin from users where id = ?")
	if err != nil {
		return false, fmt.Errorf("%s %w", op, err)
	}

	result := stmt.QueryRowContext(ctx, userID)

	var isAdmin bool
	if err = result.Scan(&isAdmin); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return isAdmin, fmt.Errorf("%s %w", op, storage.ErrNotFound)
		}

		return isAdmin, fmt.Errorf("%s %w", op, err)
	}

	return isAdmin, nil
}

func (s *Storage) App(ctx context.Context, appID int32) (models.App, error) {
	const op = "storage.postgres.app"

	stmt, err := s.db.Prepare("select * from apps where id = ?")
	if err != nil {
		return models.App{}, fmt.Errorf("%s %w", op, err)
	}

	result := stmt.QueryRowContext(ctx, appID)

	var app models.App
	if err = result.Scan(&app.Id, &app.Name, &app.Secret); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app, fmt.Errorf("%s %w", op, storage.ErrAppNotFound)
		}

		return app, fmt.Errorf("%s %w", op, err)
	}

	return app, nil
}
