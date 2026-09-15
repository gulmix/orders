package internal_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gulmix/orders/internal"
	healthController "github.com/gulmix/orders/internal/controllers/http/health"
	"github.com/gulmix/orders/internal/requestid"
)

// Тест уровня маршрутов. Он ничего не проверяет из вашего задания — он
// показывает, как тестировать HTTP-слой: через httptest, без
// поднятия настоящего сокета и без сети, поэтому такие тесты идут за
// миллисекунды и не флакают.
//
// Тесты на заказы пишете вы. Ориентир по объёму — в задании модуля.
func TestHealth(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := internal.InitRoutes(logger, healthController.NewController())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, ожидался %d", rec.Code, http.StatusOK)
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("ответ не разобрался как JSON: %v", err)
	}
	if got["status"] != "ok" {
		t.Errorf(`status = %q, ожидался "ok"`, got["status"])
	}

	// Заголовок проставляет middleware — заодно проверяем, что обвязка на месте.
	if rec.Header().Get(requestid.Header) == "" {
		t.Errorf("заголовок %s пустой", requestid.Header)
	}
}
