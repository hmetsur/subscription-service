package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"subscription-service/internal/api"
	"subscription-service/internal/config"
	"subscription-service/internal/log"
	"subscription-service/internal/repo"
	"subscription-service/internal/service"
	"syscall"
	"time"

	_ "subscription-service/docs" // Swagger docs
)

func main() {
	// 1) Конфиг + логгер
	cfg := config.MustLoad()
	logger := log.New(cfg.AppEnv)

	// 2) Подключение к БД
	db, err := repo.NewPostgres(cfg.DatabaseURL, logger)
	if err != nil {
		logger.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// 3) Миграции
	if err := repo.ApplyMigrations(db, "migrations", logger); err != nil {
		logger.Error("migrations failed", "err", err)
		os.Exit(1)
	}

	// 4) Слои
	rp := repo.NewSubscriptionsRepo(db)
	svc := service.New(rp, logger)
	h := api.NewHandlers(svc, logger)
	router := api.NewRouter(h, logger)

	// 5) HTTP-сервер
	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 6) Запуск + Graceful shutdown
	go func() {
		logger.Info("http starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http run failed", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Info("shutting down...")
	_ = srv.Shutdown(ctx)
}
