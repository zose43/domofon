package auth

import (
	"context"
	"domofon/internal/domain/models"
	"domofon/internal/storage"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (int64, error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int32) (models.App, error)
}

// NewAuth returns new instance of Auth service
func NewAuth(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidApp         = errors.New("invalid application")
)

func (a *Auth) Login(ctx context.Context, pass string, email string, appID int) (string, error) {
	const op = "auth.login"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
		slog.Int("app_id", appID),
	)

	log.Info("attempting to login user")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("user not found", err)
			return "", fmt.Errorf("%s %w", op, ErrInvalidCredentials)
		}
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", err)
			return "", fmt.Errorf("%s %w", op, ErrInvalidApp)
		}
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(pass)); err != nil {
		log.Error("invalid password", err)
		return "", fmt.Errorf("%s %w", op, ErrInvalidCredentials)
	}
	// todo handle app
}

func (a *Auth) Register(ctx context.Context, pass string, email string) (int64, error) {
	const op = "auth.register"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed generating password hash", err)
		return 0, fmt.Errorf("%s %w", op, err)
	}

	id, err := a.userSaver.SaveUser(ctx, email, hash)
	if err != nil {
		log.Error("failed saving new user", err)
		return 0, fmt.Errorf("%s %w", op, err)
	}

	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int) (bool, error) {
	panic("not implemented")
}
