package helper

import (
	"database/sql"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"net/http"
	"os"
	db "subscriptions/test/integration/db"
	"testing"
	"time"
)

var dropStatements = readFile("../../database/drop-all-tables.sql")

func readFile(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		log.Panicf("Could not read %s", file)
	}
	return string(content)
}

func ResetDatabase() {
	initIfNeeded()

	execOrPanic(dropStatements)
	driver, err := postgres.WithInstance(db.DbConnection, &postgres.Config{})
	if err != nil {
		log.Panicf("Could not create migration driver: %s", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../database/migrations",
		"postgres", driver)

	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No migrations to run, up to date")
		} else {
			log.Fatalf("Could not migrate database: %s", err)
		}
	}
}

func RunTestSetupScript(fileName string) {
	initIfNeeded()

	execOrPanic(readFile("./test-sql/" + fileName))
}

func GetDatabaseConnection() *sql.DB {
	initIfNeeded()

	return db.DbConnection
}

func ExactlyOneRowMatches(query string) error {
	var rowCount int
	if err := GetDatabaseConnection().QueryRow(query).Scan(&rowCount); err != nil {
		if err == sql.ErrNoRows {
			panic(err)
		}
	}

	if rowCount != 1 {
		return fmt.Errorf("query \"%s\" returned %d rows instead of 1", query, rowCount)
	}

	return nil
}

func AssertExactlyOneRowMatchesWithBackoff(t *testing.T, query string) {
	var check = func() error {
		return ExactlyOneRowMatches(query)
	}

	err := backoff.Retry(check, &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.2,
		MaxInterval:         1 * time.Second,
		MaxElapsedTime:      5 * time.Second,
		Stop:                -1,
		Clock:               backoff.SystemClock,
	})

	if err != nil {
		t.Fatal(err)
	}
}

func execOrPanic(statement string) {
	_, err := db.DbConnection.Exec(statement)

	if err != nil {
		log.Panicf("Could not execute database statement %s: %s", statement, err)
	}
}

func initIfNeeded() {
	if db.DbConnection == nil {
		err := db.Initialize(
			"postgres",
			"integration-test-pa55word!",
			"subscriptions",
			"localhost",
			1334)

		if err != nil {
			log.Panicf("Could not connect to database to reset for integration tests: %s", err)
		}
	}
}

func WaitForHealthcheck(t *testing.T) {
	var check = func() error {
		_, err := http.Get("http://localhost:8020/healthcheck")
		if err != nil {
			return err
		}

		return nil
	}

	err := backoff.Retry(check, &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.2,
		MaxInterval:         1 * time.Second,
		MaxElapsedTime:      30 * time.Second,
		Stop:                -1,
		Clock:               backoff.SystemClock,
	})

	if err != nil {
		t.Fatal(err)
	}
}
