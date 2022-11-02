package cashaccount

import "context"

type Storage interface {
	TopUpMoney(context.Context, *UserAmount) error
	WithdrawMoney(context.Context, *UserAmount) error
	GetAmount(context.Context, *UserID) (*UserAmount, error)
	TransferBetweenUsers(context.Context, *MoneyTransferDetails) error
	ReserveMoney(context.Context, *ReserveDetails) error
	UnreserveMoney(context.Context, *ReserveDetails) error
	// TODO report and user explanation
}
