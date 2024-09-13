package main

import (
	"log/slog"
	"net/http"
	"os"
	"zadanie-6105/config"
	"zadanie-6105/database"
	"zadanie-6105/server"
)

func main() {
	cfg, err := config.InitializeConfig()
	if err != nil {
		slog.Error("Failed to initialize config", "error", err)
		os.Exit(1)
	} else {
		slog.Info("Config initialized")
	}

	dbConnector, err := database.NewPostgresConnector(cfg)
	if err != nil {
		slog.Error("Failed to initialize database connector", "error", err)
		os.Exit(1)
	}

	runHttpServer(cfg, dbConnector)
}

func runHttpServer(cfg *config.Config, dbConnector database.DbConnector) {
	srv := server.NewServer(cfg, dbConnector)
	slog.Debug("start http server")
	router := srv.Router()
	http.Handle("/", router)
	if err := http.ListenAndServe(cfg.ServerAddress, nil); err != nil {
		slog.Warn("error during listen", "error", err)
	}
	slog.Debug("http server stopped")
}
