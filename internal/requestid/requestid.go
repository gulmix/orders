// Package requestid хранит сквозной идентификатор запроса.
//
// Отдельный маленький пакет нужен, чтобы им могли пользоваться и middleware,
// и слой ответов, не импортируя друг друга.
package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type ctxKey int

const key ctxKey = iota

// Header — заголовок, в котором идентификатор приезжает и уезжает.
// В Модуле 11 он станет частью trace context (OpenTelemetry) и будет
// пробрасываться между сервисами.
const Header = "X-Request-Id"

// FromContext возвращает идентификатор запроса или пустую строку.
func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(key).(string)
	return id
}

// NewContext кладёт идентификатор в контекст.
func NewContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, key, id)
}

// New генерирует новый идентификатор.
func New() string {
	var buf [8]byte
	// rand.Read из crypto/rand начиная с Go 1.24 не возвращает ошибку.
	_, _ = rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}
