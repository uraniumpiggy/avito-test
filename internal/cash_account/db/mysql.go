package db

import (
	"context"
	"database/sql"
	"fmt"
	"user-balance-service/internal/apperror"
	cashaccount "user-balance-service/internal/cash_account"
	"user-balance-service/pkg/logging"
)

type db struct {
	*sql.DB
	logger *logging.Logger
}

func isUserExsists(d *db, id uint32) error {
	var count int
	err := d.QueryRow(`select count(id) from service_user where id = ?;`, id).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func updateUserReport(tx *sql.Tx, user_id uint32, amount float32, description string) error {
	r, err := tx.Exec(`insert into user_report (service_user_id, amount, description) values (?, ?, ?)`, user_id, amount, description)
	if err != nil {
		return err
	}
	affected, err := r.RowsAffected()
	if affected == 0 || err != nil {
		return fmt.Errorf("Error with rows affection")
	}
	return nil
}

func (d *db) execWithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := d.BeginTx(ctx, &sql.TxOptions{Isolation: 0})
	if err != nil {
		return err
	}

	err = fn(tx)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("Rollback error")
		}
		return err
	}

	return tx.Commit()
}

func (d *db) TopUpMoney(ctx context.Context, data *cashaccount.UserAmount) error {
	err := isUserExsists(d, data.ID)
	if err != nil {
		return err
	}
	var count int
	row := d.QueryRow(`select count(id) from main_account where service_user_id = ?;`, data.ID)
	if err = row.Scan(&count); err != nil {
		return err
	}

	err = d.execWithTx(ctx, func(tx *sql.Tx) error {

		if count == 0 {
			_, err := tx.Exec(`insert into main_account (balance, service_user_id) values (?, ?);`, data.Amount, data.ID)
			if err != nil {
				return err
			}
		} else {
			_, err := tx.Exec(`update main_account set balance = balance + ? where service_user_id = ?;`, data.Amount, data.ID)
			if err != nil {
				return err
			}
		}

		err = updateUserReport(tx, data.ID, float32(data.Amount), fmt.Sprintf("Account replenished"))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		d.logger.Errorf("Error %s in topup to user: %d amount: %f", err, data.ID, data.Amount)
	} else {
		d.logger.Infof("Topup from user: %d amount: %f", err, data.ID, data.Amount)
	}
	return err
}

func (d *db) WithdrawMoney(ctx context.Context, data *cashaccount.UserAmount) error {
	err := isUserExsists(d, data.ID)
	if err != nil {
		return err
	}

	row := d.QueryRow(`select balance from main_account where service_user_id = ?;`, data.ID)

	var balance float32 = -1
	if err = row.Scan(&balance); err != nil {
		return err
	}

	err = d.execWithTx(ctx, func(tx *sql.Tx) error {
		if balance == -1 {
			return fmt.Errorf("User not have main account")
		} else {

			if balance-data.Amount < 0 {
				return fmt.Errorf("Withdraw amount is greater than balance")
			}

			_, err := tx.Exec(`update main_account set balance = balance - ? where service_user_id = ?;`, data.Amount, data.ID)
			if err != nil {
				return err
			}
		}

		err = updateUserReport(tx, data.ID, float32(data.Amount), fmt.Sprintf("Debiting money from an account"))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		d.logger.Errorf("Error %s in withdraw from user: %d amount: %f", err, data.ID, data.Amount)
	} else {
		d.logger.Infof("Withdraw from user: %d amount: %f", err, data.ID, data.Amount)
	}
	return err
}

func (d *db) GetAmount(ctx context.Context, data *cashaccount.UserID) (*cashaccount.UserAmount, error) {
	userAmount := &cashaccount.UserAmount{}

	err := isUserExsists(d, data.ID)
	if err != nil {
		return nil, err
	}

	row := d.QueryRow(`select balance from main_account where service_user_id = ?;`, data.ID)
	if err != nil {
		return nil, err
	}

	var balance float32 = -1
	if err = row.Scan(&balance); err != nil {
		return nil, err
	}

	if balance == -1 {
		return nil, fmt.Errorf("User hane not main account")
	} else {
		userAmount.Amount = balance
	}

	return userAmount, err
}

func (d *db) TransferBetweenUsers(ctx context.Context, data *cashaccount.MoneyTransferDetails) error {
	rows, err := d.Query(`select * from service_user where id in (?, ?);`, data.FromId, data.ToId)
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
		return fmt.Errorf("One or both users not found")
	}

	rows2, err := d.Query(`select service_user_id, balance from main_account where service_user_id in (?, ?);`, data.FromId, data.ToId)
	if err != nil {
		return err
	}
	defer rows2.Close()

	balances := make(map[uint32]float32)
	count = 0
	var id uint32
	var balance float32
	for rows2.Next() {
		count += 1
		err = rows2.Scan(&id, &balance)
		if err != nil {
			return err
		}
		balances[id] = balance
	}
	err = d.execWithTx(ctx, func(tx *sql.Tx) error {
		if count != 2 {
			return fmt.Errorf("One of user have not main account")
		} else {
			if balances[data.FromId] < data.Amount {
				return fmt.Errorf("User %d has insufficient funds", data.FromId)
			}
			_, err5 := tx.Exec(`update main_account set balance = balance - ? where service_user_id = ?;`, data.Amount, data.FromId)
			if err5 != nil {
				return err5
			}
			_, err5 = tx.Exec(`update main_account set balance = balance + ? where service_user_id = ?;`, data.Amount, data.ToId)
			if err5 != nil {
				return err5
			}
			err = updateUserReport(tx, data.FromId, float32(data.Amount), fmt.Sprintf("Transferring money to a user %d", data.ToId))
			if err != nil {
				return err
			}
			err = updateUserReport(tx, data.ToId, float32(data.Amount), fmt.Sprintf("Receiving money from the user %d", data.FromId))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		d.logger.Errorf("Error %s transaction from user: %d, to user: %d, amount: %f", err, data.FromId, data.ToId, data.Amount)
	} else {
		d.logger.Infof("Transaction from user: %d, to user: %d, amount: %f", data.FromId, data.ToId, data.Amount)
	}
	return err
}

func (d *db) ReserveMoney(ctx context.Context, data *cashaccount.ReserveDetails) error {
	err := isUserExsists(d, data.ID)
	if err != nil {
		return err
	}
	var balance float32
	row := d.QueryRow(`select balance from main_account where service_user_id = ?;`, data.ID)
	if err := row.Scan(&balance); err != nil {
		return err
	}
	if balance < data.Amount {
		return fmt.Errorf("User %d has insufficient funds", data.ID)
	}
	var count int
	row2 := d.QueryRow(`select count(id) from reserve_account where service_user_id = ?;`, data.ID)
	if err := row2.Scan(&count); err != nil {
		return err
	}
	err = d.execWithTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(`update main_account set balance = balance - ? where service_user_id = ?;`, data.Amount, data.ID)
		if err != nil {
			return err
		}

		if count == 0 {
			_, err := tx.Exec(`insert into reserve_account (balance, service_user_id) values (?, ?);`, data.Amount, data.ID)
			if err != nil {
				return err
			}
		} else {
			_, err := tx.Exec(`update reserve_account set balance = balance + ? where service_user_id = ?;`, data.Amount, data.ID)
			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(`insert into reservation (service_id, order_id, service_user_id, amount) values (?, ?, ?, ?);`, data.ServiceId, data.OrderId, data.ID, data.Amount)
		if err != nil {
			return err
		}

		err = updateUserReport(tx, data.ID, float32(data.Amount), fmt.Sprintf("The money %f was reserved for the order %d and the service %d", data.Amount, data.OrderId, data.ServiceId))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		d.logger.Errorf("Error %s in reserve user: %d amount: %f", err, data.ID, data.Amount)
	} else {
		d.logger.Infof("Reserve user: %d amount: %f", err, data.ID, data.Amount)
	}
	return err
}

func (d *db) AcceptRevenue(ctx context.Context, data *cashaccount.ReserveDetails) error {
	err := isUserExsists(d, data.ID)
	if err != nil {
		return err
	}

	var amount float32
	reservationRow := d.QueryRow(`select amount from reservation where service_id = ? and order_id = ? and service_user_id = ? and amount = ?;`, data.ServiceId, data.OrderId, data.ID, data.Amount)
	if err := reservationRow.Scan(&amount); err != nil {
		return err
	}

	var balance float32
	row2 := d.QueryRow(`select balance from reserve_account where service_user_id = ?;`, data.ID)
	if err := row2.Scan(&balance); err != nil {
		return err
	}
	if balance < data.Amount {
		return fmt.Errorf("Incorrect amount (not enough funds)")
	}

	err = d.execWithTx(ctx, func(tx *sql.Tx) error {
		_, err = tx.Exec(`update reserve_account set balance = balance - ? where service_user_id = ?;`, data.Amount, data.ID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`delete from reservation where service_id = ? and order_id = ? and service_user_id = ? and amount = ?`, data.ServiceId, data.OrderId, data.ID, data.Amount)
		if err != nil {
			return err
		}

		_, err := tx.Exec(`insert into bookkeeping (service_user_id, service_id, amount) values (?, ?, ?)`, data.ID, data.ServiceId, data.Amount)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		d.logger.Errorf("Error with accept money: %s", err)
		return d.execWithTx(ctx, func(tx *sql.Tx) error {
			_, err := tx.Exec(`update main_account set balance = balance + ? where service_user_id = ?;`, data.Amount, data.ID)
			if err != nil {
				return err
			}

			_, err = tx.Exec(`update reserve_account set balance = balance - ? where service_user_id = ?;`, data.Amount, data.ID)
			if err != nil {
				return err
			}

			_, err = tx.Exec(`delete from reservation where service_id = ? and order_id = ? and service_user_id = ? and amount = ?`, data.ServiceId, data.OrderId, data.ID, data.Amount)
			if err != nil {
				return err
			}

			err = updateUserReport(tx, data.ID, float32(data.Amount), fmt.Sprintf("The money %f was unreserved for the order %d and the service %d", data.Amount, data.OrderId, data.ServiceId))
			if err != nil {
				return err
			}

			return nil
		})
	}
	if err != nil {
		d.logger.Errorf("Error %s in accept user: %d, order: %d, service: %d, amount: %f", err, data.ID, data.OrderId, data.ServiceId, data.Amount)
	} else {
		d.logger.Infof("Accept user: %d, order: %d, service: %d, amount: %f", data.ID, data.OrderId, data.ServiceId, data.Amount)
	}
	return err
}

func (d *db) GetUserReport(ctx context.Context, uid, rowOffest, pageSize uint32, sortBy, sortDirection string) ([]*cashaccount.UserReportRow, error) {
	res := make([]*cashaccount.UserReportRow, 0)
	var statement string
	if rowOffest == 0 && pageSize == 0 {
		pageSize = 1000
	}
	if sortBy == "" {
		statement = `select amount, description, created_at from user_report where service_user_id = ? limit ?, ?;`
	} else {
		statement = fmt.Sprintf(`select amount, description, created_at from user_report where service_user_id = ? order by %s %s limit ?, ?;`, sortBy, sortDirection)
	}
	rows, err := d.Query(statement, uid, rowOffest, pageSize)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		item := new(cashaccount.UserReportRow)
		if err := rows.Scan(&item.Amount, &item.Description, &item.DateTime); err != nil {
			return nil, err
		}

		res = append(res, item)
	}

	return res, nil
}

func (d *db) CreateReport(ctx context.Context, timeStart, timeEnd string) ([]*cashaccount.BookkeepingReportRow, error) {
	res := make([]*cashaccount.BookkeepingReportRow, 0)
	rows, err := d.Query(`select service_id, sum(amount) as sum from bookkeeping where created_at >= ? and created_at < ? group by service_id;`, timeStart, timeEnd)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		item := new(cashaccount.BookkeepingReportRow)
		if err := rows.Scan(&item.ServiceId, &item.Amount); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (d *db) SaveReport(ctx context.Context, hash, path string) error {
	err := d.execWithTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(`delete from bookkeeping_report;`)
		_, err = tx.Exec(`insert into bookkeeping_report (hash_string, path_to_file) values (?, ?);`, hash, path)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		d.logger.Errorf("Saving bookkeeping report failed %s. Hash: %s, path: %s", err, hash, path)
	} else {
		d.logger.Info("Saved bookkeeping report. Hash: %s, path: %s", hash, path)
	}
	return err
}

func (d *db) GetReport(ctx context.Context, hash string) (string, error) {
	var path string
	r := d.QueryRow(`select path_to_file from bookkeeping_report where hash_string = ?`, hash)
	if err := r.Scan(&path); err != nil {
		return "", err
	}
	if path == "" {
		return "", apperror.ErrNotFound
	}
	return path, nil
}

func NewStorage(database *sql.DB, logger *logging.Logger) cashaccount.Storage {
	return &db{database, logger}
}
