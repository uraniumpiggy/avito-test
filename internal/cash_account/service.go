package cashaccount

import (
	"context"
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"user-balance-service/internal/apperror"
	"user-balance-service/pkg/logging"
)

type Service struct {
	storage Storage
	logger  *logging.Logger
}

func (s *Service) GetAmount(ctx context.Context, id uint32) (*UserAmount, error) {
	return s.storage.GetAmount(ctx, id)
}

func (s *Service) TopUpMoney(ctx context.Context, data *UserAmount) error {
	if data.Amount <= 0 {
		return apperror.ErrBadRequest
	}
	return s.storage.TopUpMoney(ctx, data)
}

func (s *Service) WithdrawMoney(ctx context.Context, data *UserAmount) error {
	if data.Amount <= 0 {
		return apperror.ErrBadRequest
	}
	return s.storage.WithdrawMoney(ctx, data)
}

func (s *Service) Reserve(ctx context.Context, data *ReserveDetails) error {
	if data.Amount <= 0 {
		return apperror.ErrBadRequest
	}
	if data.OrderId <= 0 || data.ServiceId <= 0 {
		return apperror.ErrBadRequest
	}
	return s.storage.ReserveMoney(ctx, data)
}

func (s *Service) AcceptRevenue(ctx context.Context, data *ReserveDetails) error {
	if data.Amount <= 0 {
		return apperror.ErrBadRequest
	}
	if data.OrderId <= 0 || data.ServiceId <= 0 {
		return apperror.ErrBadRequest
	}
	return s.storage.AcceptRevenue(ctx, data)
}

func (s *Service) TransferBetweenUsers(ctx context.Context, data *MoneyTransferDetails) error {
	if data.Amount <= 0 {
		return apperror.ErrBadRequest
	}
	if data.ToId <= 0 || data.FromId <= 0 || data.FromId == data.ToId {
		return apperror.ErrBadRequest
	}
	return s.storage.TransferBetweenUsers(ctx, data)
}

func (s *Service) GetUserReport(
	ctx context.Context,
	uid, pageNum, pageSize uint32,
	sortBy, sortDirection string) ([]*UserReportRow, error) {
	if pageNum != 0 {
		pageNum -= 1
	}
	rowOffset := pageNum * pageSize
	if sortBy != "dateTime" && sortBy != "amount" && sortBy != "" {
		return nil, apperror.ErrBadRequest
	}
	if sortDirection != "asc" && sortDirection != "desc" && sortDirection != "" {
		return nil, apperror.ErrBadRequest
	}
	if sortBy == "dateTime" {
		sortBy = "created_at"
	}
	res, err := s.storage.GetUserReport(ctx, uid, rowOffset, pageSize, sortBy, sortDirection)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) CreateBookkeepingReport(ctx context.Context, reportTime string) (string, error) {
	t, err := time.Parse("2006-01", reportTime)
	if err != nil {
		return "", err
	}
	startTime := t.Format("2006-01-02 15:04:05")
	endTime := t.AddDate(0, 1, 0).Format("2006-01-02 15:04:05")
	report, err2 := s.storage.CreateReport(ctx, startTime, endTime)
	if err2 != nil {
		return "", err2
	}

	// write to csv file
	f, err := os.Create("report.csv")
	if err != nil {
		return "", err
	}

	w := csv.NewWriter(f)
	for _, record := range report {
		row := []string{fmt.Sprintf("%d", record.ServiceId), fmt.Sprintf("%f", record.Amount)}
		if err := w.Write(row); err != nil {
			return "", err
		}
	}
	w.Flush()

	//calculate hash, define path

	path, err := os.Getwd()
	if err != nil {
		os.Remove("report.csv")
		return "", err
	}

	path = filepath.Join(path, f.Name())

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		os.Remove("report.csv")
		return "", err
	}

	hash := hex.EncodeToString(h.Sum(nil))

	//store data
	err = s.storage.SaveReport(ctx, hash, path)
	if err != nil {
		os.Remove("report.csv")
		return "", err
	}
	return hash, nil
}

func (s *Service) GetBookkeepingReport(ctx context.Context, hash string) (string, error) {
	return s.storage.GetReport(ctx, hash)
}

func NewService(st Storage, logger *logging.Logger) *Service {
	return &Service{st, logger}
}
