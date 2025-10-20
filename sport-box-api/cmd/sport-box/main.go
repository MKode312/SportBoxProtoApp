package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	bookgrpc "sport-box-api/internal/clients/booking/grpc"
	paymgrpc "sport-box-api/internal/clients/payments/grpc"
	ssogrpc "sport-box-api/internal/clients/sso/grpc"
	"sport-box-api/internal/config"
	"sport-box-api/internal/http-server/handlers/auth/login"
	"sport-box-api/internal/http-server/handlers/auth/register"
	"sport-box-api/internal/http-server/handlers/book"
	"sport-box-api/internal/http-server/handlers/paym/addcard"
	"sport-box-api/internal/http-server/handlers/paym/addfunds"
	authMW "sport-box-api/internal/http-server/middleware/auth"
	mwLogger "sport-box-api/internal/http-server/middleware/logger"
	"sport-box-api/internal/lib/logger/handlers/slogpretty"
	"sport-box-api/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envlocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("config", slog.Any("cfg", cfg))

	os.Setenv("APP_SECRET", cfg.AppSecret)

	log.Info("server running", slog.String("env", cfg.Env))

	ssoClient, err := ssogrpc.New(
		context.Background(),
		log,
		cfg.Clients.SSO.Addr,
		cfg.Clients.SSO.Timeout,
		cfg.Clients.SSO.RetriesCount,
	)
	if err != nil {
		log.Error("failed to init sso client", sl.Err(err))
		os.Exit(1)
	}

	paymentsClient, err := paymgrpc.New(
		context.Background(),
		log,
		cfg.Clients.Payments.Addr,
		cfg.Clients.Payments.Timeout,
		cfg.Clients.Payments.RetriesCount,
	)
	if err != nil {
		log.Error("failed to init payments client", sl.Err(err))
		os.Exit(1)
	}

	bookingClient, err := bookgrpc.New(
		context.Background(),
		log,
		cfg.Clients.Booking.Addr,
		cfg.Clients.Booking.Timeout,
		cfg.Clients.Booking.RetriesCount,
	)
	if err != nil {
		log.Error("failed to init booking client", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/register", register.New(context.Background(), log, *ssoClient))
	router.Post("/login", login.New(context.Background(), log, *ssoClient))

	router.Group(func(r chi.Router) {
		r.Use(authMW.AuthorizeJWTToken)
		r.Post("/addcard", addcard.New(context.Background(), log, *paymentsClient))
		r.Post("/addfunds", addfunds.New(context.Background(), log, *paymentsClient))
		r.Post("/book", book.New(context.Background(), log, *bookingClient))
	})

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
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
