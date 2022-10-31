package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func NewClient(ctx context.Context, host, port, username, password, database string) (*sql.DB, error) {
	dbURL := fmt.Sprintf("jdbc:mysql://%s:%s@%s:%s/%s", username, password, host, port, database)
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		return nil, errors.New("Failed to connect to database")
	}

	err = db.Ping()
	if err != nil {
		return nil, errors.New("Failed to ping database")
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(10)

	return db, nil
}
