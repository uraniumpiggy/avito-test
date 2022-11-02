package cashaccount

import (
	"encoding/json"
	"net/http"
	"user-balance-service/internal/handlers"
	"user-balance-service/pkg/logging"

	"github.com/julienschmidt/httprouter"
)

type handler struct {
	service *Service
	logger  *logging.Logger
}

func NewHandler(service *Service, logger *logging.Logger) handlers.Handler {
	return &handler{
		service: service,
		logger:  logger,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.PATCH("/api/accrual/", h.Accrual)
	router.PATCH("/api/reserve/", h.Accrual)
	router.PATCH("/api/accept_transfer/", h.Accrual)
	router.GET("/api/user_balance/", h.GetUserBalance)
	router.GET("/api/report/", h.Accrual)
	router.GET("/api/user_transactions/", h.Accrual)
}

func (h *handler) Accrual(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

}

func (h *handler) GetUserBalance(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var uid UserID
	err := json.NewDecoder(r.Body).Decode(&uid)
	if err != nil {

	}

	userAmount, err := h.service.GetAmount(r.Context(), &uid)
	h.logger.Info(userAmount)
	if err != nil {

	}

	json.NewEncoder(w).Encode(userAmount)
}
