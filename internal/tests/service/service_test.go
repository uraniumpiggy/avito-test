package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"
	cashaccount "user-balance-service/internal/cash_account"
	"user-balance-service/internal/cash_account/db"
	"user-balance-service/pkg/client/mysql"
	"user-balance-service/pkg/logging"
)

var (
	d *sql.DB
	s *cashaccount.Service
)

func prepareService() {
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
	service := cashaccount.NewService(storage, logger)

	d = database
	s = service
}

func TestMain(t *testing.M) {
	prepareService()
	d.Exec(`delete from main_account;`)
	d.Exec(`delete from reservation;`)
	d.Exec(`delete from reserve_account;`)
	d.Exec(`delete from user_report;`)
	d.Exec(`delete from bookkeeping;`)
	t.Run()
	os.RemoveAll("all.log")
	os.Exit(0)
}

func TestGetUserReport(t *testing.T) {
	err := s.TopUpMoney(context.Background(), &cashaccount.UserAmount{ID: 1, Amount: 100})
	time.Sleep(1 * time.Second)
	err2 := s.TopUpMoney(context.Background(), &cashaccount.UserAmount{ID: 2, Amount: 100})
	err3 := s.TransferBetweenUsers(context.Background(), &cashaccount.MoneyTransferDetails{FromId: 1, ToId: 2, Amount: 40})
	time.Sleep(1 * time.Second)
	err4 := s.WithdrawMoney(context.Background(), &cashaccount.UserAmount{ID: 1, Amount: 10})
	time.Sleep(1 * time.Second)
	err5 := s.Reserve(context.Background(), &cashaccount.ReserveDetails{ID: 1, ServiceId: 1, OrderId: 1, Amount: 20})
	time.Sleep(1 * time.Second)
	err6 := s.AcceptRevenue(context.Background(), &cashaccount.ReserveDetails{ID: 1, ServiceId: 1, OrderId: 1, Amount: 20})

	if err != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil {
		panic("Cant prepare data")
	}

	urr, err := s.GetUserReport(context.Background(), uint32(1), 0, 0, "", "")
	if err != nil {
		t.Error(err)
	}

	if len(urr) != 5 {
		t.Errorf("Incorrect len of array %d must be 4", len(urr))
	}

	compareArray := []*cashaccount.UserReportRow{
		{
			Amount:      100,
			Description: "Account replenished",
		},
		{
			Amount:      40,
			Description: "Transferring money to a user 2",
		},
		{
			Amount:      10,
			Description: "Debiting money from an account",
		},
		{
			Amount:      20,
			Description: "The money 20.000000 was reserved for the order 1 and the service 1",
		},
		{
			Amount:      20,
			Description: "The money 20.000000 was accepted for the order 1 and the service 1",
		},
	}

	for i := 0; i < len(urr); i++ {
		if urr[i].Amount != compareArray[i].Amount || urr[i].Description != compareArray[i].Description {
			t.Error("Wrong array")
		}
	}

	urr1, err := s.GetUserReport(context.Background(), uint32(1), 1, 3, "", "")
	if err != nil {
		t.Error(err)
	}

	if len(urr1) != 3 {
		t.Error(len(urr1))
	}

	for i := 0; i < len(urr1); i++ {
		if urr1[i].Amount != compareArray[i].Amount || urr1[i].Description != compareArray[i].Description {
			t.Error("Wrong array")
		}
	}

	urr2, err := s.GetUserReport(context.Background(), uint32(1), 1, 100, "dateTime", "desc")
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < len(urr2); i++ {
		if urr2[i].Amount != compareArray[len(compareArray)-1-i].Amount || urr2[i].Description != compareArray[len(compareArray)-1-i].Description {
			t.Error("Wrong array")
		}
	}

	urr3, err := s.GetUserReport(context.Background(), uint32(1), 1, 3, "amount", "desc")
	if err != nil {
		t.Error(err)
	}

	for i := 1; i < len(urr3); i++ {
		fmt.Println(urr3[i])
		if urr3[i].Amount >= urr3[i-1].Amount {
			t.Error("wrong array")
		}
	}

	urr4, err := s.GetUserReport(context.Background(), uint32(1), 1, 5, "amount", "asc")
	if err != nil {
		t.Error(err)
	}

	if len(urr4) != 5 {
		t.Error(len(urr4))
	}

	for i := 1; i < len(urr4); i++ {
		if urr4[i].Amount < urr4[i-1].Amount {
			t.Error("wrong array")
		}
	}

	urr4, err = s.GetUserReport(context.Background(), uint32(1), 1, 5, "a", "a")
	if err == nil {
		t.Error("Err must me not nil")
	}

	urr4, err = s.GetUserReport(context.Background(), uint32(100), 1, 5, "", "")

	if len(urr4) != 0 {
		t.Error(len(urr4))
	}

	d.Exec(`delete from main_account;`)
	d.Exec(`delete from reservation;`)
	d.Exec(`delete from reserve_account;`)
	d.Exec(`delete from user_report;`)
	d.Exec(`delete from bookkeeping;`)

}
