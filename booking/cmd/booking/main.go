package main

import (
	"booking/internal/app"
	"booking/internal/clients/payments"
	"booking/internal/config"
	"booking/internal/lib/logger/handlers/slogpretty"
	"booking/internal/lib/logger/sl"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envlocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("app running")

	paymentsClient, err := payments.New(context.Background(), log, cfg.Clients.Payments.Address, cfg.Clients.Payments.Timeout, cfg.Clients.Payments.RetriesCount)
	if err != nil {
		log.Error("failed to start payments client", sl.Err(err))
		os.Exit(1)
	}

	application := app.New(log, *paymentsClient, cfg.GRPC.Addr, cfg.StoragePath)

	go application.GRPCSrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("stopping application", slog.String("signal", sign.String()))

	application.GRPCSrv.Stop()

	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envlocal:
		log = setupPretySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPretySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
