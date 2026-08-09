// Package internal собирает маршруты приложения в одном месте.
//
// Контроллеры лежат по пакетам, а карта URL — здесь: по этому файлу видно
// весь публичный интерфейс сервиса, не открывая двадцать пакетов.
package internal

import (
	"log/slog"
	"net/http"

	healthController "github.com/gulmix/orders/internal/controllers/http/health"
	"github.com/gulmix/orders/internal/middlewares"
)

// InitRoutes собирает маршруты и оборачивает их общей обвязкой.
//
// Используется стандартный http.ServeMux: с Go 1.22 он умеет метод и
// wildcard в паттерне («POST /orders», «GET /orders/{id}»), поэтому внешний
// роутер на онбординге не нужен. Значение wildcard достаётся из запроса
// через r.PathValue("id").
func InitRoutes(logger *slog.Logger, health *healthController.Controller) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", health.Health)

	// Порядок обёрток важен: request id должен появиться раньше логирования,
	// а recover — охватывать всё остальное.
	var h http.Handler = mux
	h = middlewares.Logging(logger, h)
	h = middlewares.Recover(logger, h)
	h = middlewares.RequestID(h)

	return h
}
