package grpcapp

import (
	"fmt"
	"log/slog"
	"net"
	authgrpc "sso/internal/grpc/auth"

	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	addr      string
}

func New(log *slog.Logger, authService authgrpc.Auth, addr string) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer, authService)

	return &App{
		log: log,
		gRPCServer: gRPCServer,
		addr: addr,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.String("addr", a.addr),
	)

	l, err := net.Listen("tcp", a.addr)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("starting gRPC server", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).Info("stopping gRPC server", slog.String("addr", a.addr))

	a.gRPCServer.GracefulStop()
}