package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"user-balance-service/pkg/logging"

	_ "github.com/go-sql-driver/mysql"
)

func NewClient(ctx context.Context, host, port, username, password, database string) (*sql.DB, error) {
	logger := logging.NewLogger()
	dbURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", username, password, host, port, database)
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database due to error %s", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Failed to ping to database due to error %s", err)
	}

	logger.Info("Successfully connected to database")

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(20)

	return db, nil
}
