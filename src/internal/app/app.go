package app

import (
	grpcapp "domofon/internal/app/grpc"
	"domofon/internal/services/auth"
	"domofon/internal/storage/postgres"
	"log/slog"
	"time"
)

type App struct {
	GrpcSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	storageUrl string,
	port int,
	tokenTTL time.Duration,
) *App {
	storage, err := postgres.NewStorage(storageUrl)
	if err != nil {
		panic(err)
	}

	authService := auth.NewAuth(log, storage, storage, storage, tokenTTL)

	return &App{GrpcSrv: grpcapp.New(
		log,
		port,
		authService,
	)}
}
