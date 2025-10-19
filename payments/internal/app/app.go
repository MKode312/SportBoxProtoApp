package app

import (
	"log/slog"
	"payments/internal/services/payment"
	"payments/internal/storage/sqlite"
	grpcApp "payments/internal/app/grpc"
)

type App struct {
	GRPCSrv *grpcApp.App
}

func New(log *slog.Logger, gRPCAddr string, storagePath string) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	paymentService := payment.New(log, storage, storage, storage)

	grpcApp := grpcApp.New(log, paymentService, gRPCAddr)

	return &App{
		GRPCSrv: grpcApp,
	}
}
