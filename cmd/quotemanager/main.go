package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"quotemanager/internal/config"
	"quotemanager/internal/handlers"
	"quotemanager/internal/repositories"
	"time"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", ".env", "configuration file")
	flag.Parse()

	cfg := config.MustLoadCfg(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	log.Info("starting server")

	log.Debug("debug messages are enabled")

	// db

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBConfig.DBHost, cfg.DBConfig.DBUser, cfg.DBConfig.DBPassword, cfg.DBConfig.DBName, cfg.DBConfig.DBPort)

	storage, err := repositories.New(log, dsn)
	if err != nil {
		log.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}
	if err := storage.Migrate(); err != nil {
		log.Error("failed to migrate db", "error", err)
		os.Exit(1)
	}

	log.Info("successfully connected to database")

	mux := http.NewServeMux()

	mux.Handle("POST /quotes", handlers.AddQuoteHandler(log, storage))
	mux.Handle("GET /quotes", handlers.GetQuotesHandler(log, storage))
	mux.Handle("GET /quotes/random", handlers.GetRandomQuoteHandler(log, storage))
	mux.Handle("DELETE /quotes/{quoteID}", handlers.DeleteQuoteHandler(log, storage))

	server := http.Server{
		Addr:        cfg.HttpServerAddress,
		ReadTimeout: cfg.HttpServerTimeout * time.Second,
		Handler:     mux,
	}

	log.Info("server is listening on", "address", cfg.HttpServerAddress)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("erroneous shutdown", "error", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server closed unexpectedly", "error", err)
			return
		}
	}

}

func mustMakeLogger(logLevel string) *slog.Logger {
	return slog.Default()
}
