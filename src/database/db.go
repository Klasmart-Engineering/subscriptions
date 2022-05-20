package db

import (
	"database/sql"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	_ "github.com/lib/pq"
	"log"
)

const (
	PORT = 5432
)

// ErrNoMatch is returned when we request a row that doesn't exist
var ErrNoMatch = fmt.Errorf("no matching record")

type Database struct {
	Conn *sql.DB
}

func Initialize(username, password, database, host string) (Database, error) {
	db := Database{}

	connect := func() error {
		log.Println("Attempting to connect to database")
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, PORT, username, password, database)
		conn, err := sql.Open("postgres", dsn)
		if err != nil {
			return err
		}
		db.Conn = conn
		err = db.Conn.Ping()
		if err != nil {
			return err
		}
		log.Println("Database connection established")
		return nil
	}

	err := backoff.Retry(connect, backoff.NewExponentialBackOff())

	return db, err
}
