package storage

import (
	"context"
	"database/sql"
	"os"
	"testing"
	cashaccount "user-balance-service/internal/cash_account"
	"user-balance-service/internal/cash_account/db"
	"user-balance-service/pkg/client/mysql"
	"user-balance-service/pkg/logging"
)

var s cashaccount.Storage
var d *sql.DB

func prepareStorage() {
	logger := logging.NewLogger()

	database, err := mysql.NewClient(
		context.Background(),
		"localhost",
		"3306",
		"user",
		"secret",
		"service-db")
	if err != nil {
		panic(err)
	}

	storage := db.NewStorage(database, logger)

	s = storage
	d = database
}

func TestMain(t *testing.M) {
	prepareStorage()
	t.Run()
	os.RemoveAll("all.log")
	os.Exit(0)
}

func TestTopUpMoney(t *testing.T) {
	d.Exec(`delete from main_account;`)
	d.Exec(`delete from reservation;`)
	d.Exec(`delete from reserve_account;`)
	data := &cashaccount.UserAmount{
		ID:     999,
		Amount: 100,
	}
	err := s.TopUpMoney(context.Background(), data)
	if err == nil {
		t.Error()
	}

	data.ID = 1
	data.Amount = -10
	err = s.TopUpMoney(context.Background(), data)
	if err == nil {
		t.Error()
	}

	data.ID = 1
	data.Amount = 10.32
	err = s.TopUpMoney(context.Background(), data)
	if err != nil {
		t.Error()
	}
	var balance float32
	r := d.QueryRow(`select balance from main_account where service_user_id = ?`, data.ID)
	if err = r.Scan(&balance); err != nil {
		t.Error()
	}

	if balance != data.Amount {
		t.Error(balance)
	}

	d.Exec(`delete from main_account where service_user_id = ?;`, data.ID)
}

func TestWithdraw(t *testing.T) {
	err := s.TopUpMoney(context.Background(), &cashaccount.UserAmount{ID: 1, Amount: 100})
	if err != nil {
		panic(err)
	}

	data := &cashaccount.UserAmount{
		ID:     999,
		Amount: 100,
	}
	err = s.WithdrawMoney(context.Background(), data)
	if err == nil {
		t.Error()
	}

	data.ID = 1
	data.Amount = 200
	err = s.WithdrawMoney(context.Background(), data)
	if err == nil {
		t.Error()
	}

	data.Amount = 20
	err = s.WithdrawMoney(context.Background(), data)
	if err != nil {
		t.Error()
	}

	var balance float32
	r := d.QueryRow(`select balance from main_account where service_user_id = ?`, data.ID)
	if err = r.Scan(&balance); err != nil {
		t.Error(err)
	}

	if balance != 80 {
		t.Error(balance)
	}

	d.Exec(`delete from main_account where service_user_id = ?;`, data.ID)
}

func TestGetBalance(t *testing.T) {
	err := s.TopUpMoney(context.Background(), &cashaccount.UserAmount{ID: 1, Amount: 100})
	if err != nil {
		panic(err)
	}

	var data uint32 = 100
	_, err = s.GetAmount(context.Background(), data)
	if err == nil {
		t.Error()
	}

	data = 1
	ua, err := s.GetAmount(context.Background(), 1)
	if err != nil {
		t.Error(err)
	}
	if ua.Amount != 100 {
		t.Error(ua.Amount)
	}

	d.Exec(`delete from main_account where service_user_id = ?;`, data)
}

func TestTransactionBetweenUsers(t *testing.T) {
	err := s.TopUpMoney(context.Background(), &cashaccount.UserAmount{ID: 1, Amount: 100})
	if err != nil {
		panic(err)
	}
	err = s.TopUpMoney(context.Background(), &cashaccount.UserAmount{ID: 2, Amount: 100})
	if err != nil {
		panic(err)
	}

	data := &cashaccount.MoneyTransferDetails{
		FromId: 1,
		ToId:   100,
		Amount: 10,
	}

	err = s.TransferBetweenUsers(context.Background(), data)
	if err == nil {
		t.Error("Transfer to unexisted user")
	}

	data.ToId = 2
	data.Amount = 200
	err = s.TransferBetweenUsers(context.Background(), data)
	if err == nil {
		t.Error("Transfer too much money")
	}

	data.Amount = 20
	err = s.TransferBetweenUsers(context.Background(), data)
	if err != nil {
		t.Error(err)
	}

	var balance1, balance2 float32
	r := d.QueryRow(`select balance from main_account where service_user_id = ?`, data.ToId)
	if err = r.Scan(&balance1); err != nil {
		t.Error(err)
	}

	if balance1 != 120 {
		t.Error(balance1)
	}

	r = d.QueryRow(`select balance from main_account where service_user_id = ?`, data.FromId)
	if err = r.Scan(&balance2); err != nil {
		t.Error(err)
	}

	if balance2 != 80 {
		t.Error(balance2)
	}

	d.Exec(`delete from main_account where service_user_id = ?;`, data.ToId)
	d.Exec(`delete from main_account where service_user_id = ?;`, data.FromId)
}

func TestReserve(t *testing.T) {
	err := s.TopUpMoney(context.Background(), &cashaccount.UserAmount{ID: 1, Amount: 100})
	if err != nil {
		panic(err)
	}

	data := &cashaccount.ReserveDetails{
		ID:        100,
		ServiceId: 1,
		OrderId:   1,
		Amount:    10,
	}
	err = s.ReserveMoney(context.Background(), data)
	if err == nil {
		t.Error("User not exist")
	}

	data.ID = 1
	data.Amount = 200
	err = s.ReserveMoney(context.Background(), data)
	if err == nil {
		t.Error("Amount too much")
	}

	data.Amount = 10
	err = s.ReserveMoney(context.Background(), data)
	if err != nil {
		t.Error(err)
	}
	var balance float32
	r := d.QueryRow(`select balance from main_account where service_user_id = ?`, data.ID)
	if err = r.Scan(&balance); err != nil {
		t.Error(err)
	}

	if balance != 90 {
		t.Error(balance)
	}

	r = d.QueryRow(`select balance from reserve_account where service_user_id = ?`, data.ID)
	if err = r.Scan(&balance); err != nil {
		t.Error(err)
	}

	if balance != 10 {
		t.Error(balance)
	}
	d.Exec(`delete from main_account where service_user_id = ?;`, data.ID)
	d.Exec(`delete from reserve_account where service_user_id = ?;`, data.ID)
	d.Exec(`delete from reservation where service_user_id = ?;`, data.ID)

}

func TestAcceptRevenue(t *testing.T) {
	err := s.TopUpMoney(context.Background(), &cashaccount.UserAmount{ID: 1, Amount: 100})
	if err != nil {
		panic(err)
	}

	data := &cashaccount.ReserveDetails{
		ID:        1,
		ServiceId: 1,
		OrderId:   1,
		Amount:    10,
	}
	err = s.ReserveMoney(context.Background(), data)
	if err != nil {
		panic(err)
	}

	data.ID = 100
	err = s.AcceptRevenue(context.Background(), data)
	if err == nil {
		t.Error("Accept from unexisted user")
	}

	data.ID = 1
	data.ServiceId = 2
	err = s.AcceptRevenue(context.Background(), data)
	if err == nil {
		t.Error("Accept with wrong service_id")
	}

	data.ServiceId = 1
	data.ID = 1
	data.Amount = 2000
	err = s.AcceptRevenue(context.Background(), data)
	if err == nil {
		t.Error("Accept with wrong amount")
	}

	data.Amount = 10
	err = s.AcceptRevenue(context.Background(), data)
	if err != nil {
		t.Error(err)
	}

	var balance float32
	r := d.QueryRow(`select balance from main_account where service_user_id = ?`, data.ID)
	if err = r.Scan(&balance); err != nil {
		t.Error(err)
	}

	if balance != 90 {
		t.Error(balance)
	}

	r = d.QueryRow(`select balance from reserve_account where service_user_id = ?`, data.ID)
	if err = r.Scan(&balance); err != nil {
		t.Error(err)
	}

	if balance != 0 {
		t.Error(balance)
	}
}
