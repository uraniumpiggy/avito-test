package cashaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	router.PATCH("/api/withdraw/", h.Withdraw)
	router.PATCH("/api/reserve/", h.Reserve)
	router.PATCH("/api/accept_transfer/", h.AcceptTransfer)
	router.PATCH("/api/users_transfer/", h.UsersTransfer)
	router.GET("/api/user_balance/", h.GetUserBalance)
	router.POST("/api/report_create/", h.CreateReport)
	router.GET("/api/report/:hash", h.GetReport)
	router.GET("/api/user_report/", h.GetUserReport)
}

func (h *handler) Accrual(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var data UserAmount
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	err = h.service.TopUpMoney(context.Background(), &data)
	if err != nil {
		w.WriteHeader(400)
		return
	}
}

func (h *handler) Withdraw(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var data UserAmount
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	err = h.service.WithdrawMoney(context.Background(), &data)
	if err != nil {
		w.WriteHeader(400)
		return
	}
}

func (h *handler) GetUserBalance(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var uid UserID
	err := json.NewDecoder(r.Body).Decode(&uid)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	userAmount, err := h.service.GetAmount(context.Background(), &uid)
	h.logger.Info(userAmount)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	json.NewEncoder(w).Encode(userAmount)
}

func (h *handler) Reserve(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var details ReserveDetails
	err := json.NewDecoder(r.Body).Decode(&details)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	err = h.service.Reserve(context.Background(), &details)
	if err != nil {
		w.WriteHeader(400)
		return
	}
}

func (h *handler) AcceptTransfer(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var details ReserveDetails
	err := json.NewDecoder(r.Body).Decode(&details)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	err = h.service.AcceptRevenue(context.Background(), &details)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}
}

func (h *handler) UsersTransfer(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var transferData MoneyTransferDetails
	err := json.NewDecoder(r.Body).Decode(&transferData)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	err = h.service.TransferBetweenUsers(context.Background(), &transferData)
	if err != nil {
		w.WriteHeader(400)
		return
	}
}

func (h *handler) GetUserReport(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	urlId := r.URL.Query().Get("id")
	urlPageNum := r.URL.Query().Get("pageNum")
	urlPageSize := r.URL.Query().Get("pageSize")
	sortBy := r.URL.Query().Get("sortBy")
	sortDirection := r.URL.Query().Get("sortDirection")
	if urlId == "" {
		w.WriteHeader(400)
		return
	}
	uid, err := strconv.Atoi(urlId)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	if urlPageNum == "" && urlPageSize != "" || urlPageNum != "" && urlPageSize == "" {
		w.WriteHeader(400)
		return
	}

	if sortBy == "" && sortDirection != "" || sortBy != "" && sortDirection == "" {
		w.WriteHeader(400)
		return
	}

	var pageNum, pageSize int
	if urlPageNum != "" {
		pageNum, err = strconv.Atoi(urlPageNum)
		if err != nil {
			w.WriteHeader(400)
			return
		}
	}

	if urlPageSize != "" {
		pageSize, err = strconv.Atoi(urlPageSize)
		if err != nil {
			w.WriteHeader(400)
			return
		}
	}

	urr, err := h.service.GetUserReport(context.Background(), uint32(uid), uint32(pageNum), uint32(pageSize), sortBy, sortDirection)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	json.NewEncoder(w).Encode(urr)
}

func (h *handler) CreateReport(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	startTime := r.URL.Query().Get("startTime")
	if startTime == "" {
		w.WriteHeader(400)
		return
	}
	hash, err := h.service.CreateBookkeepingReport(context.Background(), startTime)
	fmt.Println(err)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	fmt.Println(hash)
}

func (h *handler) GetReport(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hash := params.ByName("hash")

	path, err := h.service.GetBookkeepingReport(context.Background(), hash)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=report.csv")
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, path)
}
