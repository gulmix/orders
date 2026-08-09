// Package response — единый формат ответа для всех HTTP-контроллеров.
package response

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gulmix/orders/internal/requestid"
)

// Единый формат ошибки для всего API — часть контракта, менять его нельзя.
// Клиент всегда получает один и тот же конверт, поэтому его можно разбирать
// машинно, а не глазами:
//
//	{"error": {"code": "validation_error", "message": "...", "request_id": "..."}}
//
// В Модуле 4 вы добавите сюда поле details и версионирование контракта.
type errorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

// Коды ошибок — тоже часть публичного контракта: клиент ветвится по ним,
// а не по тексту сообщения. Менять их можно только через депрекацию.
const (
	// CodeBadRequest — запрос невозможно разобрать (например, битый JSON).
	CodeBadRequest = "bad_request"
	// CodeValidation — запрос разобран, но нарушает правила предметной области.
	CodeValidation = "validation_error"
	// CodeNotFound — запрошенного ресурса не существует.
	CodeNotFound = "not_found"
	// CodeInternal — ошибка на стороне сервиса. Подробности наружу не отдаём.
	CodeInternal = "internal_error"
)

// JSON отправляет успешный ответ.
func JSON(w http.ResponseWriter, r *http.Request, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// Заголовки уже ушли — ответ починить нельзя, остаётся только зафиксировать.
		slog.ErrorContext(r.Context(), "не удалось закодировать ответ", slog.Any("error", err))
	}
}

// Error отправляет ошибку в едином формате.
//
// Сообщение попадает к клиенту: пишите то, что ему поможет, и не выносите
// наружу внутренние подробности — для них есть лог.
func Error(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	JSON(w, r, status, errorResponse{Error: errorBody{
		Code:      code,
		Message:   message,
		RequestID: requestid.FromContext(r.Context()),
	}})
}

// Перевод доменных ошибок в HTTP-коды — ваша часть работы, и делать её стоит
// в одном месте, а не в каждом хендлере: домен не должен знать про 404,
// а контроллеры — про устройство ошибок домена. Разбирать ошибку следует через
// errors.Is / errors.As, чтобы обёртки fmt.Errorf("...: %w", err) не ломали
// логику. Набросок, который вы допишете сами в controllers/http/order:
//
//	func writeDomainError(w http.ResponseWriter, r *http.Request, err error) {
//		switch {
//		case errors.Is(err, order.ErrNotFound):
//			response.Error(w, r, http.StatusNotFound, response.CodeNotFound, "заказ не найден")
//		// ...остальные случаи...
//		default:
//			slog.ErrorContext(r.Context(), "необработанная ошибка", slog.Any("error", err))
//			response.Error(w, r, http.StatusInternalServerError, response.CodeInternal, "внутренняя ошибка сервиса")
//		}
//	}
