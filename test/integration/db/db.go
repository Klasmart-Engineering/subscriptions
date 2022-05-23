package integration_test_db

import (
	"database/sql"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	_ "github.com/lib/pq"
	"log"
	"time"
)

// ErrNoMatch is returned when we request a row that doesn't exist
var ErrNoMatch = fmt.Errorf("no matching record")

type Database struct {
	Conn *sql.DB
}

func Initialize(username, password, database, host string, port int) (Database, error) {
	db := Database{}

	connect := func() error {
		log.Printf("Attempting to connect to database %s@%s:%d", username, host, port)
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, username, password, database)
		conn, err := sql.Open("postgres", dsn)
		if err != nil {
			log.Printf("Could not connect to database: %s", err)
			return err
		}
		db.Conn = conn
		err = db.Conn.Ping()
		if err != nil {
			log.Printf("Could not ping database: %s", err)
			return err
		}

		log.Println("Connected to database")
		return nil
	}
	err := backoff.Retry(connect, &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.2,
		MaxInterval:         5 * time.Second,
		MaxElapsedTime:      60 * time.Second,
		Stop:                -1,
		Clock:               backoff.SystemClock,
	})

	return db, err
}
