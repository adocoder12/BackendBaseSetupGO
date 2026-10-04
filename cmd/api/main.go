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

	"github.com/adocoder12/BackendBaseSetupGO/internal/config"
	"github.com/adocoder12/BackendBaseSetupGO/internal/db"
	"github.com/adocoder12/BackendBaseSetupGO/internal/handlers"
)

func main() {
	// 1. Create the logger (prints JSON to the terminal)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	// Make it the default logger for the whole app
	slog.SetDefault(logger)

	// 2. Run the app. If it returns an error, log it and exit.
	// Only main calls os.Exit, so cleanup inside run() always happens first.
	if err := run(logger); err != nil {
		logger.Error("fatal", "error", err)
		os.Exit(1)
	}
}

// run holds the real startup logic so deferred cleanup (pool.Close, cancel)
// always runs. os.Exit in main would skip it.
func run(logger *slog.Logger) error {
	// 3. Load settings (reads the .env file, falls back to defaults)
	config, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// 4. Listen for stop signals (Ctrl+C, docker stop, Air restart).
	// When one arrives, ctx is cancelled.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 5. Connect to the database. Give up if it takes more than 5 seconds.
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := db.NewPool(dbCtx, config.Database)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	// Close the pool when run() ends (runs last, after the server stops)
	defer pool.Close()
	logger.Info("database connection established")

	// 6. Apply any new migrations
	if err := db.Migrate(config.Database); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	logger.Info("database migrations verified")

	// 7. Build the app: handlers + routes
	app := handlers.NewApplication(logger)

	// 8. Configure the HTTP server
	server := &http.Server{
		Addr:         ":" + config.Server.Port,
		Handler:      app.SetupRoutes(),
		IdleTimeout:  time.Minute,      // close idle connections after 1 min
		ReadTimeout:  5 * time.Second,  // max time to read a request
		WriteTimeout: 10 * time.Second, // max time to send a response
	}

	// 9. Start the server in the background (goroutine).
	// ListenAndServe blocks, so without "go" run() would be stuck here.
	// If the server fails, the error is sent through the channel.
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("API running", "port", config.Server.Port)
		// ErrServerClosed is the normal result of Shutdown, not a real error
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// 10. Wait here until one of these happens
	select {
	case err := <-serverErr: // the server crashed
		return fmt.Errorf("server failed: %w", err)
	case <-ctx.Done(): // Ctrl+C / docker stop / Air restart
		logger.Info("shutting down")
	}

	// 11. Graceful shutdown: stop taking new requests, let running ones finish.
	// Waits up to 10 seconds, then cuts off any request still running.
	// Uses a fresh context because ctx is already cancelled.
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	// When this returns, run() ends and the deferred calls run:
	// pool.Close() first, then cancel() and stop()
	return server.Shutdown(shutdownCtx)
}
