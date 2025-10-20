package app

import (
	grpcapp "booking/internal/app/grpc"
	"booking/internal/clients/payments"
	"booking/internal/lib/logger/sl"
	"booking/internal/services/book"
	"booking/internal/storage/sqlite"
	"context"
	"log/slog"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(ctx context.Context, log *slog.Logger, paymclient payments.Client, interval int64, grpcAddr string, storagePath string) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}
	errCh := storage.StartDbChecker(ctx, interval)

	go func() {
		for err := range errCh {
			if err != nil {
				log.Error("db check worker error", sl.Err(err))
			}
		}
	}()

	bookingService := book.NewBooker(log, storage)

	grpcApp := grpcapp.New(log, bookingService, paymclient, grpcAddr)

	return &App{
		GRPCSrv: grpcApp,
	}
}