package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/adocoder12/BackendBaseSetupGO/internal/config"
	"github.com/adocoder12/BackendBaseSetupGO/internal/db"
	"github.com/adocoder12/BackendBaseSetupGO/internal/handlers"
	"github.com/joho/godotenv"
)

func main() {
	//settings our logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
	//get env

	if err := godotenv.Load(); err != nil {
		logger.Warn("no .env file found")
	}
	//config load
	config, err := config.Load()
	if err != nil {
		logger.Error("failed load config", "error", err)
		os.Exit(1)

	}
	// db
	pool, err := db.NewPool(config.Database)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("database connection stablished!")

	// Run migrations
	if err := db.Migrate(config.Database); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("database migrations verified")

	app := handlers.NewApplication(logger)
	router := app.SetupRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	logger.Info(" Costa PMS API running on", "port", config.Server.Port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}

}
