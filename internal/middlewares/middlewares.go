// Package middlewares — общая обвязка HTTP-запросов: идентификатор запроса,
// структурированный лог, защита от паники.
package middlewares

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gulmix/orders/internal/controllers/http/response"
	"github.com/gulmix/orders/internal/requestid"
)

// RequestID берёт идентификатор из заголовка или генерирует новый и кладёт
// его в контекст: всё, что происходит внутри запроса, логируется с этим id.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestid.Header)
		if id == "" {
			id = requestid.New()
		}

		w.Header().Set(requestid.Header, id)
		next.ServeHTTP(w, r.WithContext(requestid.NewContext(r.Context(), id)))
	})
}

// Logging пишет по одной структурированной строке на запрос.
// Структурированный лог — требование курса: логи читает машина, а не человек.
func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		logger.LogAttrs(r.Context(), slog.LevelInfo, "http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.Duration("duration", time.Since(start)),
			slog.String("request_id", requestid.FromContext(r.Context())),
		)
	})
}

// Recover не даёт паникующему хендлеру уронить весь процесс.
func Recover(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.ErrorContext(r.Context(), "паника в хендлере",
					slog.Any("panic", rec),
					slog.String("request_id", requestid.FromContext(r.Context())),
				)
				response.Error(w, r, http.StatusInternalServerError, response.CodeInternal, "внутренняя ошибка сервиса")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// statusRecorder запоминает код ответа, чтобы его можно было залогировать.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.status = status
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(status)
}
