package auth

import (
	"context"
	"domofon/internal/domain/models"
	"domofon/internal/lib/jwt"
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
	ErrUserExists         = errors.New("user already exists")
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
		log.Error("failed getting user by email", err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(pass)); err != nil {
		log.Error("invalid password", err)
		return "", fmt.Errorf("%s %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, int32(appID))
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Error("app not found")
			return "", fmt.Errorf("%s %s", op, ErrInvalidApp)
		}
		log.Error("failed getting app by app_id", err)
	}

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed generating token", err)
		return "", fmt.Errorf("%s %w", op, err)
	}

	return token, nil
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
		if errors.Is(err, storage.ErrUserExists) {
			log.Error("user already exists")
			return 0, fmt.Errorf("%s %w", op, ErrUserExists)
		}
		log.Error("failed saving new user", err)
	}

	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int) (bool, error) {
	const op = "auth.isAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("user_id", userID),
	)

	result, err := a.userProvider.IsAdmin(ctx, int64(userID))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Error("user not found")
			return result, fmt.Errorf("%s %w", op, ErrInvalidCredentials)
		}
		log.Error("failed check admin status")
	}

	return result, nil
}
