package grpcapp

import (
	"domofon/internal/grpc/auth"
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type App struct {
	log     *slog.Logger
	grpcSrv *grpc.Server
	port    int
}

func New(
	log *slog.Logger,
	port int,
	authService auth.Auth,
) *App {
	grpcSrv := grpc.NewServer()
	auth.Register(grpcSrv, authService)

	return &App{log: log, port: port, grpcSrv: grpcSrv}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(slog.String("op", op))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("starting grpc Server", slog.String("addr", l.Addr().String()))

	if err := a.grpcSrv.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	slog.With(slog.String("op", op)).
		Info("stopping grpc Server")

	a.grpcSrv.GracefulStop()
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}
