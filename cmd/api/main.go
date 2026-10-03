package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

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
	logger.Info("Server running")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}

}
