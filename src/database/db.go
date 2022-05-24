package db

import (
	"database/sql"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"subscriptions/src/monitoring"
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
		monitoring.GlobalContext.Info("Attempting to connect to database",
			zap.String("username", username),
			zap.String("host", host),
			zap.Int("port", port))
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, username, password, database)
		conn, err := sql.Open("postgres", dsn)
		if err != nil {
			monitoring.GlobalContext.Error("Could not connect to database", zap.Error(err))
			return err
		}
		db.Conn = conn
		err = db.Conn.Ping()
		if err != nil {
			monitoring.GlobalContext.Error("Could not ping database", zap.Error(err))
			return err
		}

		monitoring.GlobalContext.Info("Database connection established")
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
