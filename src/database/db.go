package db

import (
	"database/sql"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	_ "github.com/newrelic/go-agent/v3/integrations/nrpq"
	"go.uber.org/zap"
	"subscriptions/src/monitoring"
	"time"
)

var dbConnection *sql.DB

func Initialize(username, password, database, host string, port int) error {
	connect := func() error {
		monitoring.GlobalContext.Info("Attempting to connect to database",
			zap.String("username", username),
			zap.String("host", host),
			zap.Int("port", port))
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, username, password, database)
		conn, err := sql.Open("nrpostgres", dsn)
		if err != nil {
			monitoring.GlobalContext.Error("Could not connect to database", zap.Error(err))
			return err
		}
		dbConnection = conn
		err = dbConnection.Ping()
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

	if err != nil {
		return err
	}

	migrateDatabase(dbConnection)

	return nil
}

func Close() {
	dbConnection.Close()
}
