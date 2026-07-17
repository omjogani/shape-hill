package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/omjogani/shape-hill/internal/config"
	"github.com/omjogani/shape-hill/internal/server"
	"github.com/omjogani/shape-hill/internal/store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("shapehill exited", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	log := setupLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := openStore(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	return serve(ctx, newServer(cfg, db, log), log)
}

func loadConfig() (*config.Config, error) {
	return config.Load(".")
}

func setupLogger(level string) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel(level)}))
}

func openStore(ctx context.Context, databaseURL string) (*store.Store, error) {
	return store.New(ctx, databaseURL)
}

func newServer(cfg *config.Config, db *store.Store, log *slog.Logger) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           server.New(db, log),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func serve(ctx context.Context, srv *http.Server, log *slog.Logger) error {
	// A failed ListenAndServe cancels this context too, so a startup error (e.g.
	// port in use) unblocks the wait below instead of hanging forever.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		log.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "err", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func logLevel(name string) slog.Level {
	switch name {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
