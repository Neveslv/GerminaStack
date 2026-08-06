package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"germinaStack/auth"
	"germinaStack/config"
	"germinaStack/database"
	"germinaStack/handlers"
	"germinaStack/routes"
)

const (
	tokenTTL          = 24 * time.Hour
	migrationTimeout  = 15 * time.Second
	shutdownTimeout   = 10 * time.Second
	httpReadTimeout   = 10 * time.Second
	httpResponseGrace = 2 * time.Second
)

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}

func main() {
	if err := run(); err != nil {
		log.Print("application terminated due to an initialization or runtime error")
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return errors.New("invalid application configuration")
	}

	rootContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(rootContext, cfg.DatabaseURL, database.DefaultPoolConfig())
	if err != nil {
		return errors.New("database initialization failed")
	}
	defer db.Close()

	migrationContext, cancelMigration := context.WithTimeout(rootContext, migrationTimeout)
	defer cancelMigration()
	if err := database.Migrate(migrationContext, db); err != nil {
		return errors.New("database migration failed")
	}

	mailer, err := auth.NewSMTPMailer(cfg.SMTP)
	if err != nil {
		return errors.New("SMTP initialization failed")
	}
	credentials := database.NewPostgresCredentialRepository(db)
	challenges := database.NewPostgresChallengeRepository(db)
	catalog := database.NewPostgresCatalogRepository(db)
	discussion := database.NewPostgresDiscussionRepository(db)
	account := database.NewPostgresAccountRepository(db)
	authService := auth.NewService(credentials, challenges, mailer, []byte(cfg.TwoFactorSecret), systemClock{})
	authHandler := handlers.NewAuthHandler(authService, cfg.JWTSecret, cfg.CookieSecure, tokenTTL, cfg.AuthOperationTimeout)
	userHandler := handlers.NewUserHandler(credentials, cfg.AuthOperationTimeout)
	catalogHandler := handlers.NewCatalogHandler(catalog, cfg.AuthOperationTimeout)
	discussionHandler := handlers.NewDiscussionHandler(discussion, cfg.AuthOperationTimeout)
	accountHandler := handlers.NewAccountHandler(account, cfg.AuthOperationTimeout)
	router := routes.NewRouter(authHandler, userHandler, catalogHandler, discussionHandler, accountHandler, cfg.JWTSecret)

	server := newHTTPServer(cfg.HTTPAddr, router, cfg.AuthOperationTimeout)
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return errors.New("HTTP server failed")
	case <-rootContext.Done():
		shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancelShutdown()
		if err := server.Shutdown(shutdownContext); err != nil {
			return errors.New("HTTP server shutdown failed")
		}
		return nil
	}
}

func newHTTPServer(address string, handler http.Handler, operationTimeout time.Duration) *http.Server {
	writeTimeout := operationTimeout +
		httpReadTimeout +
		auth.ChallengeInvalidationTimeout +
		httpResponseGrace

	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       httpReadTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       60 * time.Second,
	}
}
