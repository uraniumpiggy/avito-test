package cashaccount

import "context"

type Storage interface {
	TopUpMoney(context.Context, *UserAmount) error
	WithdrawMoney(context.Context, *UserAmount) error
	GetAmount(context.Context, uint32) (*UserAmount, error)
	TransferBetweenUsers(context.Context, *MoneyTransferDetails) error
	ReserveMoney(context.Context, *ReserveDetails) error
	AcceptRevenue(ctx context.Context, data *ReserveDetails) error
	GetUserReport(ctx context.Context, uid, rowOffest, pageSize uint32, sortBy, sortDirection string) ([]*UserReportRow, error)
	CreateReport(ctx context.Context, timeStart, timeEnd string) ([]*BookkeepingReportRow, error)
	SaveReport(ctx context.Context, hash, path string) error
	GetReport(ctx context.Context, hash string) (string, error)
}
