package cashaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"user-balance-service/internal/apperror"
	"user-balance-service/internal/config"
	"user-balance-service/internal/handlers"
	"user-balance-service/internal/middleware"
	"user-balance-service/pkg/logging"

	"github.com/julienschmidt/httprouter"
)

type handler struct {
	service *Service
	logger  *logging.Logger
	cache   map[string]string
}

func NewHandler(service *Service, logger *logging.Logger) handlers.Handler {
	return &handler{
		service: service,
		logger:  logger,
		cache:   make(map[string]string),
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, "/api/users/accrual/", middleware.Middleware(h.Accrual))
	router.HandlerFunc(http.MethodPost, "/api/users/withdraw/", middleware.Middleware(h.Withdraw))
	router.HandlerFunc(http.MethodPost, "/api/users/reserve/", middleware.Middleware(h.Reserve))
	router.HandlerFunc(http.MethodPost, "/api/users/accept/", middleware.Middleware(h.AcceptTransfer))
	router.HandlerFunc(http.MethodPost, "/api/users/transaction/", middleware.Middleware(h.UsersTransfer))
	router.HandlerFunc(http.MethodGet, "/api/users/balance/:id", middleware.Middleware(h.GetUserBalance))
	router.HandlerFunc(http.MethodPost, "/api/report/create/", middleware.Middleware(h.CreateReport))
	router.HandlerFunc(http.MethodGet, "/api/report/:hash", middleware.Middleware(h.GetReport))
	router.HandlerFunc(http.MethodGet, "/api/users/report/", middleware.Middleware(h.GetUserReport))
}

func (h *handler) Accrual(w http.ResponseWriter, r *http.Request) error {
	var data UserAmount
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return err
	}

	err = h.service.TopUpMoney(context.Background(), &data)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) Withdraw(w http.ResponseWriter, r *http.Request) error {
	var data UserAmount
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return apperror.ErrBadRequest
	}

	err = h.service.WithdrawMoney(context.Background(), &data)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) GetUserBalance(w http.ResponseWriter, r *http.Request) error {
	params := httprouter.ParamsFromContext(r.Context())
	userId := params.ByName("id")

	numUserId, err := strconv.Atoi(userId)
	if err != nil {
		return apperror.ErrBadRequest
	}

	userAmount, err := h.service.GetAmount(context.Background(), uint32(numUserId))
	h.logger.Info(userAmount)
	if err != nil {
		return err
	}

	json.NewEncoder(w).Encode(userAmount)

	return nil
}

func (h *handler) Reserve(w http.ResponseWriter, r *http.Request) error {
	var details ReserveDetails
	err := json.NewDecoder(r.Body).Decode(&details)
	if err != nil {
		return apperror.ErrBadRequest
	}

	err = h.service.Reserve(context.Background(), &details)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) AcceptTransfer(w http.ResponseWriter, r *http.Request) error {
	var details ReserveDetails
	err := json.NewDecoder(r.Body).Decode(&details)
	if err != nil {
		return apperror.ErrBadRequest
	}

	err = h.service.AcceptRevenue(context.Background(), &details)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) UsersTransfer(w http.ResponseWriter, r *http.Request) error {
	var transferData MoneyTransferDetails
	err := json.NewDecoder(r.Body).Decode(&transferData)
	if err != nil {
		return apperror.ErrBadRequest
	}

	err = h.service.TransferBetweenUsers(context.Background(), &transferData)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) GetUserReport(w http.ResponseWriter, r *http.Request) error {
	urlId := r.URL.Query().Get("id")
	urlPageNum := r.URL.Query().Get("pageNum")
	urlPageSize := r.URL.Query().Get("pageSize")
	sortBy := r.URL.Query().Get("sortBy")
	sortDirection := r.URL.Query().Get("sortDirection")
	if urlId == "" {
		return apperror.ErrBadRequest
	}
	uid, err := strconv.Atoi(urlId)
	if err != nil {
		return err
	}

	if urlPageNum == "" && urlPageSize != "" || urlPageNum != "" && urlPageSize == "" {
		return apperror.ErrBadRequest
	}

	if sortBy == "" && sortDirection != "" || sortBy != "" && sortDirection == "" {
		return apperror.ErrBadRequest
	}

	var pageNum, pageSize int
	if urlPageNum != "" {
		pageNum, err = strconv.Atoi(urlPageNum)
		if err != nil {
			return err
		}
	}

	if urlPageSize != "" {
		pageSize, err = strconv.Atoi(urlPageSize)
		if err != nil {
			return err
		}
	}

	if pageNum < 1 && urlPageNum != "" || pageSize < 1 && urlPageSize != "" {
		return apperror.ErrBadRequest
	}

	urr, err := h.service.GetUserReport(context.Background(), uint32(uid), uint32(pageNum), uint32(pageSize), sortBy, sortDirection)
	if err != nil {
		return err
	}

	json.NewEncoder(w).Encode(urr)

	return nil
}

func (h *handler) CreateReport(w http.ResponseWriter, r *http.Request) error {
	startTime := r.URL.Query().Get("startTime")
	if startTime == "" {
		return apperror.ErrBadRequest
	}
	hash, err := h.service.CreateBookkeepingReport(context.Background(), startTime)
	if err != nil {
		return err
	}

	cfg := config.GetConfig()
	resp := &ReportLinkResponse{
		Link: fmt.Sprintf("%s:%s/api/report/%s", cfg.Listen.BindIp, cfg.Listen.Port, hash),
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(resp)
	return nil
}

func (h *handler) GetReport(w http.ResponseWriter, r *http.Request) error {
	params := httprouter.ParamsFromContext(r.Context())
	hash := params.ByName("hash")
	if hash == "" {
		return apperror.ErrBadRequest
	}

	path, err := h.service.GetBookkeepingReport(context.Background(), hash)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Disposition", "attachment; filename=report.csv")
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, path)
	return nil
}
