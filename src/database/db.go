package db

import (
	"database/sql"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/jmoiron/sqlx"
	_ "github.com/newrelic/go-agent/v3/integrations/nrpq"
	"go.uber.org/zap"
	"regexp"
	"strings"
	"subscriptions/src/monitoring"
	"time"
)

var dbConnection *sqlx.DB

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

		dbConnection = sqlx.NewDb(conn, "nrpostgres")
		dbConnection.MapperFunc(FromPascalCaseToSnakeCase)
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

	migrateDatabase(dbConnection.DB)
	seedDatabase(dbConnection.DB)

	return nil
}

func Close() {
	dbConnection.Close()
}

func FromPascalCaseToSnakeCase(sqlFieldName string) string {
	var s = sqlFieldName
	for _, reStr := range []string{`([A-Z]+)([A-Z][a-z])`, `([a-z\d])([A-Z])`} {
		re := regexp.MustCompile(reStr)
		s = re.ReplaceAllString(s, "${1}_${2}")
	}

	return strings.ToLower(s)
}
