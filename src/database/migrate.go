package db

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"subscriptions/src/monitoring"
)

func migrateDatabase() {
	monitoring.GlobalContext.Info("Starting database migration")

	driver, err := postgres.WithInstance(dbConnection.DB, &postgres.Config{})
	if err != nil {
		monitoring.GlobalContext.Fatal("Could not create migration driver", zap.Error(err))
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./database/migrations",
		"postgres", driver)

	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			monitoring.GlobalContext.Info("No migrations to run, up to date")
		} else {
			monitoring.GlobalContext.Fatal("Could not migrate database", zap.Error(err))
		}
	}

	monitoring.GlobalContext.Info("Finished database migration")
}
