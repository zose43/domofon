package auth

import (
	"context"
	"domofon/internal/domain/models"
	"log"
	"time"
)

type Auth struct {
	log          *log.Logger
	userSaver    UserProvider
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (int64, error)
}

type UserProvider interface {
	UserSaver(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int32) (models.App, error)
}

// NewAuth returns new instance of Auth service
func NewAuth(
	log *log.Logger,
	userSaver UserProvider,
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

func (a *Auth) Login(ctx context.Context, pass string, email string, appID int) (string, error) {
	panic("not implemented")
}

func (a *Auth) Register(ctx context.Context, pass string, email string) (int64, error) {
	panic("not implemented")
}

func (a *Auth) IsAdmin(ctx context.Context, userID int) (bool, error) {
	panic("not implemented")
}
