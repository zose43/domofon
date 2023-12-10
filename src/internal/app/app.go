package app

import (
	grpcapp "domofon/internal/app/grpc"
	"log/slog"
	"time"
)

type App struct {
	GrpcSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	storagePath string,
	port int,
	tokenTTL time.Duration,
) *App {
	return &App{GrpcSrv: grpcapp.New(log, port)}
}
