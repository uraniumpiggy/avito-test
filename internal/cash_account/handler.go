package cashaccount

import (
	"net/http"
	"user-balance-service/internal/handlers"
	"user-balance-service/pkg/logging"

	"github.com/julienschmidt/httprouter"
)

type handler struct {
	logger *logging.Logger
}

func NewHandler(logger *logging.Logger) handlers.Handler {
	return &handler{
		logger: logger,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.PATCH("/api/accrual/", h.Accrual)
	router.PATCH("/api/reserve/", h.Accrual)
	router.PATCH("/api/accept_transfer/", h.Accrual)
	router.GET("/api/user_balance/", h.Accrual)
	router.GET("/api/report/", h.Accrual)
	router.GET("/api/user_transactions/", h.Accrual)
}

func (h *handler) Accrual(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

}
