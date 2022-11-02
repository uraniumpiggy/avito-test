package cashaccount

import (
	"context"
	"user-balance-service/pkg/logging"
)

type Service struct {
	storage Storage
	logger  *logging.Logger
}

func (s *Service) GetAmount(ctx context.Context, data *UserID) (*UserAmount, error) {
	return s.storage.GetAmount(ctx, data)
}

func (s *Service) TopUpMoney(ctx context.Context, data *UserAmount) error {
	return s.storage.TopUpMoney(ctx, data)
}

func NewService(st Storage, logger *logging.Logger) *Service {
	return &Service{st, logger}
}
