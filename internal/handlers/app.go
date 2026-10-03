package handlers

import "log/slog"

type App struct {
	logger *slog.Logger
}

// constructor
func NewApplication(logger *slog.Logger) *App {
	return &App{logger: logger}
}
