package main

import (
	"context"
	"fmt"
	"net/http"
	cashaccount "user-balance-service/internal/cash_account"
	"user-balance-service/internal/cash_account/db"
	"user-balance-service/internal/config"
	"user-balance-service/pkg/client/mysql"
	"user-balance-service/pkg/logging"

	"github.com/julienschmidt/httprouter"
)

func main() {
	logger := logging.NewLogger()

	logger.Info("Craete router")
	router := httprouter.New()

	cfg := config.GetConfig()

	database, err := mysql.NewClient(
		context.Background(),
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Database)
	if err != nil {
		panic(err)
	}

	logger.Info("Creating storage")
	storage := db.NewStorage(database, logger)

	logger.Info("Creating service")
	service := cashaccount.NewService(storage, logger)

	logger.Info("Register handler")
	handler := cashaccount.NewHandler(service, logger)

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
