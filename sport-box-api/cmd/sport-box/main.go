package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	bookgrpc "sport-box-api/internal/clients/booking/grpc"
	paymgrpc "sport-box-api/internal/clients/payments/grpc"
	ssogrpc "sport-box-api/internal/clients/sso/grpc"
	"sport-box-api/internal/config"
	"sport-box-api/internal/http-server/handlers/auth/login"
	"sport-box-api/internal/http-server/handlers/auth/register"
	"sport-box-api/internal/http-server/handlers/book"
	"sport-box-api/internal/http-server/handlers/paym/addcard"
	"sport-box-api/internal/http-server/handlers/paym/addfunds"
	"sport-box-api/internal/http-server/handlers/paym/getcard"
	authMW "sport-box-api/internal/http-server/middleware/auth"
	mwLogger "sport-box-api/internal/http-server/middleware/logger"
	"sport-box-api/internal/lib/logger/handlers/slogpretty"
	"sport-box-api/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

	// Configure CORS
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080", "http://localhost:8082", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposedHeaders:   []string{"Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	// Serve static frontend files
	frontendPath := http.Dir("/app/sport-box-frontend")
	FileServer(router, "/", frontendPath)

	// API routes
	router.Route("/api", func(r chi.Router) {
		// Public routes
		r.Post("/auth/register", register.New(context.Background(), log, *ssoClient))
		r.Post("/auth/login", login.New(context.Background(), log, *ssoClient))

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMW.AuthorizeJWTToken)
			r.Post("/payments/add-card", addcard.New(context.Background(), log, *paymentsClient))
			r.Post("/payments/add-funds", addfunds.New(context.Background(), log, *paymentsClient))
			r.Get("/payments/cards", getcard.New(context.Background(), log, *paymentsClient))
			r.Post("/book", book.New(context.Background(), log, *bookingClient))
			r.Get("/boxes", book.GetBoxes(context.Background(), log, *bookingClient))
			r.Get("/bookings", book.GetBookings(context.Background(), log, *bookingClient))
			r.Delete("/bookings/{id}", book.Cancel(bookingClient))
		})
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

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
