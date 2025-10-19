package app

import (
	grpcapp "booking/internal/app/grpc"
	"booking/internal/clients/payments"
	"booking/internal/services/book"
	"booking/internal/storage/sqlite"
	"log/slog"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, paymclient payments.Client, grpcAddr string, storagePath string) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	bookingService := book.NewBooker(log, storage)


	grpcApp := grpcapp.New(log, bookingService, paymclient, grpcAddr)

	return &App{
		GRPCSrv: grpcApp,
	}
}
