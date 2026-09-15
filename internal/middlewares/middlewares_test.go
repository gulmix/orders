package middlewares_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gulmix/orders/internal/middlewares"
)

// chain повторяет порядок обёрток из InitRoutes: тесты ниже проверяют
// цепочку целиком, а не каждую обёртку по отдельности, потому что ломается
// обычно именно порядок. Меняете порядок в routes.go — меняйте и здесь,
// иначе тесты начнут проверять не то, что работает в сервисе.
func chain(logger *slog.Logger, next http.Handler) http.Handler {
	h := middlewares.Recover(logger, next)
	h = middlewares.Logging(logger, h)
	return middlewares.RequestID(h)
}

// Паникующий хендлер — это тоже запрос, и в логе он должен выглядеть как
// обычный: метод, путь, статус 500. Recover снаружи Logging такую строку
// теряет — логировать становится некому, паника не доходит до Logging живой.
func TestПаникаОтдаётся500ИПопадаетВЛог(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	h := chain(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("бум")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/orders", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, ожидался %d", rec.Code, http.StatusInternalServerError)
	}

	var body struct {
		Error struct {
			Code      string `json:"code"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("ответ не разобрался как JSON: %v", err)
	}
	if body.Error.Code != "internal_error" {
		t.Errorf("code = %q, ожидался \"internal_error\"", body.Error.Code)
	}
	if body.Error.RequestID == "" {
		t.Error("request_id в ответе пустой")
	}

	var access map[string]any
	for line := range strings.Lines(logs.String()) {
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry["msg"] == "http request" {
			access = entry
		}
	}
	if access == nil {
		t.Fatalf("в логе нет строки запроса, только: %s", logs.String())
	}
	if status, _ := access["status"].(float64); int(status) != http.StatusInternalServerError {
		t.Errorf("в логе status = %v, ожидался 500", access["status"])
	}
	if access["request_id"] != body.Error.RequestID {
		t.Errorf("request_id в логе (%v) и в ответе (%q) разошлись", access["request_id"], body.Error.RequestID)
	}
}

// Обвязка подменяет ResponseWriter своей обёрткой, и хендлер не должен из-за
// этого лишиться стриминга: http.NewResponseController добирается до Flush
// через метод Unwrap.
func TestОбвязкаНеЛомаетFlush(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))

	var flushErr error
	h := chain(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("часть ответа"))
		flushErr = http.NewResponseController(w).Flush()
	}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/stream", nil))

	if flushErr != nil {
		t.Errorf("Flush через ResponseController: %v", flushErr)
	}
}

// http.ErrAbortHandler хендлер кидает намеренно, чтобы оборвать ответ.
// Recover обязан пробросить её дальше: гасит такую панику сам net/http,
// и превращать её в 500 нельзя.
func TestОбрывОтветаПробрасывается(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))

	h := chain(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	defer func() {
		rec := recover()
		if rec == nil {
			t.Fatal("паника не проброшена: Recover проглотил ErrAbortHandler")
		}
		err, ok := rec.(error)
		if !ok || !errors.Is(err, http.ErrAbortHandler) {
			t.Fatalf("проброшено %v, ожидался http.ErrAbortHandler", rec)
		}
	}()

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/orders/1", nil))
}
