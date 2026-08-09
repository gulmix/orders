// Package health — liveness-проба и одновременно образец контроллера: так
// выглядит ответ, прошедший через response.JSON и общую обвязку.
package health

import (
	"net/http"

	"github.com/gulmix/orders/internal/controllers/http/response"
)

// Controller держит зависимости хендлеров пробы. Их пока нет.
type Controller struct{}

// NewController создаёт контроллер.
func NewController() *Controller {
	return &Controller{}
}

// Health отвечает, что процесс жив.
func (c *Controller) Health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
}
