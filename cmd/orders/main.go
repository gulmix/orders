// Команда orders — HTTP-сервис заказов, сквозной проект курса.
//
// main держит только composition root: читает конфигурацию, собирает
// зависимости, запускает сервер и корректно его останавливает. Логики здесь
// нет и быть не должно — её невозможно протестировать.
package main

import (
	"log/slog"
	"os"

	"github.com/gulmix/orders/internal"
	"github.com/gulmix/orders/internal/config"
	healthController "github.com/gulmix/orders/internal/controllers/http/health"
	httpServer "github.com/gulmix/orders/internal/servers/http"
)

func main() {
	if err := run(); err != nil {
		slog.Error("сервис остановлен с ошибкой", slog.Any("error", err))
		os.Exit(1)
	}
}

// run возвращает ошибку вместо того, чтобы звать os.Exit по всему коду:
// так путь завершения один и его видно целиком.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	// Composition root: зависимости создаются здесь и передаются явно, снизу
	// вверх — хранилище в сервис, сервис в контроллер, контроллер в маршруты.
	// Никаких глобальных переменных и пакетов-синглтонов.
	health := healthController.NewController()

	srv := httpServer.NewServer(cfg.HTTPAddr, internal.InitRoutes(logger, health), logger, cfg.ShutdownTimeout)

	return srv.Run()
}
