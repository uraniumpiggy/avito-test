package db

import (
	"context"
	"database/sql"
	"fmt"
	cashaccount "user-balance-service/internal/cash_account"
	"user-balance-service/pkg/logging"
)

type db struct {
	*sql.DB
	logger *logging.Logger
}

func (d *db) execWithTx(ctx context.Context, fn func() error) error {
	tx, err := d.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	err = fn()

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

func (d *db) TopUpMoney(ctx context.Context, data *cashaccount.UserAmount) error {
	err := d.execWithTx(ctx, func() error {
		stmt, err := d.Prepare(`select count(id) from service_user where id = ?;`)
		stmt2, err2 := d.Prepare(`select count(m.id) as count from service_user as u, main_account as m where m.service_user_id = ?;`)
		stmt3, err3 := d.Prepare(`insert into main_account (balance, service_user_id) values (?, ?);`)
		stmt4, err4 := d.Prepare(`update main_account set balance = balance + ? where service_user_id = ?;`)

		if err != nil || err2 != nil || err3 != nil || err4 != nil {
			return fmt.Errorf("Errors in prepared statements %s, %s, %s, %s", err, err2, err3, err4)
		}

		defer stmt.Close()
		defer stmt2.Close()
		defer stmt3.Close()
		defer stmt4.Close()

		rows, err := stmt.Query(data.ID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var count int
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}
		if count == 0 {
			return fmt.Errorf("User not found")
		}

		rows2, err := stmt2.Query(data.ID)
		if err != nil {
			return err
		}
		defer rows2.Close()

		for rows2.Next() {
			err = rows2.Scan(&count)
			if err != nil {
				return err
			}
		}

		if count == 0 { // user not have main_account lets create it
			_, err := stmt3.Exec(data.Amount, data.ID)
			if err != nil {
				return err
			}
		} else { // user have main_account we ned top up it
			_, err := stmt4.Exec(data.Amount, data.ID)
			if err != nil {
				return err
			}
		}

		return nil
	})
	return err
}

func (d *db) WithdrawMoney(ctx context.Context, data *cashaccount.UserAmount) error {
	err := d.execWithTx(ctx, func() error {
		stmt, err := d.Prepare(`select count(id) from service_user where id = ?;`)
		stmt2, err2 := d.Prepare(`select balance from main_account where service_user_id = ?;`)
		stmt4, err4 := d.Prepare(`update main_account set balance = balance - ? where service_user_id = ?;`)

		if err != nil || err2 != nil || err4 != nil {
			return fmt.Errorf("Errors in prepared statements %s, %s, %s", err, err2, err4)
		}

		defer stmt.Close()
		defer stmt2.Close()
		defer stmt4.Close()

		rows, err := stmt.Query(data.ID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var count int
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}
		if count == 0 {
			return fmt.Errorf("User not found")
		}

		rows2, err := stmt2.Query(data.ID)
		if err != nil {
			return err
		}
		defer rows2.Close()

		var balance int = -1
		for rows2.Next() {
			err = rows2.Scan(&balance)
			if err != nil {
				return err
			}
		}

		if balance == -1 { // user not have main_account
			// TODO
		} else { // user have main_account withdraw it

			if balance-int(data.Amount) < 0 {
				return fmt.Errorf("Withdraw amount is greater than balance")
			}

			_, err := stmt4.Exec(data.Amount, data.ID)
			if err != nil {
				return err
			}
		}

		return nil
	})
	return err
}

func (d *db) GetAmount(ctx context.Context, data *cashaccount.UserID) (*cashaccount.UserAmount, error) {
	userAmount := &cashaccount.UserAmount{}

	err := d.execWithTx(ctx, func() error {
		stmt, err := d.Prepare(`select count(id) from service_user where id = ?;`)
		stmt2, err2 := d.Prepare(`select balance from main_account where service_user_id = ?;`)

		if err != nil || err2 != nil {
			return fmt.Errorf("Errors in prepared statements %s, %s", err, err2)
		}

		defer stmt.Close()
		defer stmt2.Close()

		rows, err := stmt.Query(data.ID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var count int
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}
		if count == 0 {
			return fmt.Errorf("User not found")
		}

		rows2, err := stmt2.Query(data.ID)
		if err != nil {
			return err
		}
		defer rows2.Close()

		var balance int = -1
		for rows2.Next() {
			err = rows2.Scan(&balance)
			if err != nil {
				return err
			}
		}

		if balance == -1 { // user not have main_account
			return fmt.Errorf("User hane not main account")
		} else { // user have main_account withdraw it
			userAmount.Amount = uint32(balance)
		}

		return nil
	})

	return userAmount, err
}

func (d *db) TransferBetweenUsers(ctx context.Context, data *cashaccount.MoneyTransferDetails) error {
	err := d.execWithTx(ctx, func() error {
		stmt, err := d.Prepare(`select * from service_user where id in (?, ?);`)
		stmt2, err2 := d.Prepare(`select service_user_id, balance from main_account where service_user_id in (?, ?);`)
		stmt3, err3 := d.Prepare(`update main_account set balance = balance - ? where service_user_id = ?;`)
		stmt4, err4 := d.Prepare(`update main_account set balance = balance + ? where service_user_id = ?;`)

		if err != nil || err2 != nil || err3 != nil || err4 != nil {
			return fmt.Errorf("Errors in prepared statements %s, %s, %s, %s", err, err2, err3, err4)
		}

		defer stmt.Close()
		defer stmt2.Close()
		defer stmt3.Close()
		defer stmt4.Close()

		rows, err := stmt.Query(data.FromId, data.ToId)
		if err != nil {
			return err
		}
		defer rows.Close()

		var count int
		for rows.Next() {
			count += 1
			if err != nil {
				return err
			}
		}
		if count != 2 {
			return fmt.Errorf("One of users not found") // !!! fix it
		}

		rows2, err := stmt2.Query(data.FromId, data.ToId)
		if err != nil {
			return err
		}
		defer rows2.Close()

		balances := make(map[uint32]uint32)
		count = 0
		var id, balance uint32
		for rows2.Next() {
			count += 1
			err = rows2.Scan(&id, &balance)
			if err != nil {
				return err
			}
			balances[id] = balance
		}

		if count != 2 { // user not have main_account lets create it
			return fmt.Errorf("One of user have not main account")
		} else {
			if balances[data.FromId] < data.Amount {
				return fmt.Errorf("Not enough money")
			}
			_, err5 := stmt3.Exec(data.Amount, data.FromId)
			if err5 != nil {
				return err5
			}
			_, err5 = stmt4.Exec(data.Amount, data.ToId)
			if err5 != nil {
				return err5
			}
		}

		return nil
	})
	return err
}

func (d *db) ReserveMoney(ctx context.Context, data *cashaccount.ReserveDetails) error {
	return nil
}

func (d *db) UnreserveMoney(ctx context.Context, data *cashaccount.ReserveDetails) error {
	return nil
}

func NewStorage(database *sql.DB, logger *logging.Logger) cashaccount.Storage {
	return &db{database, logger}
}
