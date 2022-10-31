package main

import (
	"fmt"
	"net/http"
	cashaccount "user-balance-service/internal/cash_account"
	"user-balance-service/internal/config"
	"user-balance-service/pkg/logging"

	"github.com/julienschmidt/httprouter"
)

func main() {
	logger := logging.NewLogger()

	logger.Info("Craete router")
	router := httprouter.New()

	cfg := config.GetConfig()

	logger.Info("Register handler")
	handler := cashaccount.NewHandler(logger)

	handler.Register(router)
	start(router, cfg)
}

func start(router *httprouter.Router, cfg *config.Config) {
	logger := logging.NewLogger()

	logger.Infof("Start application on %s:%s", cfg.Listen.BindIp, cfg.Listen.Port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", cfg.Listen.BindIp, cfg.Listen.Port), router)
	if err != nil {
		panic(fmt.Errorf("Can't start server due to err: %s", err))
	}
}
