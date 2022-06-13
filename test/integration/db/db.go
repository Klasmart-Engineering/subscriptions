package integration_test_db

import (
	"database/sql"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	_ "github.com/lib/pq"
	"log"
	"time"
)

var DbConnection *sql.DB

func Initialize(username, password, database, host string, port int) error {
	connect := func() error {
		log.Printf("Attempting to connect to database %s@%s:%d", username, host, port)
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, username, password, database)
		conn, err := sql.Open("postgres", dsn)
		if err != nil {
			log.Printf("Could not connect to database: %s", err)
			return err
		}
		DbConnection = conn
		err = DbConnection.Ping()
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

	return err
}
