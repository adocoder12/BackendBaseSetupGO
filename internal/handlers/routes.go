package handlers

import (
	"fmt"
	"net/http"
)

func (app *App) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /home", func(w http.ResponseWriter, r *http.Request) {
		app.logger.Info("Home handler")
		_, _ = fmt.Fprint(w, "hello from home, good ado!!\n")
	})
	return mux
}
