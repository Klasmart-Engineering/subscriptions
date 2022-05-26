package helper

import (
	"database/sql"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"log"
	"net/http"
	"os"
	db "subscriptions/test/integration/db"
	"testing"
	"time"
)

var connection *db.Database
var dropStatements = readFile("../../database/drop-all-tables.sql")
var initStatements = readFile("../../database/init.sql")

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
	execOrPanic(initStatements)
}

func RunTestSetupScript(fileName string) {
	initIfNeeded()

	execOrPanic(readFile("./test-sql/" + fileName))
}

func GetDatabaseConnection() *db.Database {
	initIfNeeded()

	return connection
}

func ExactlyOneRowMatches(query string) error {
	var rowCount int
	if err := GetDatabaseConnection().Conn.QueryRow(query).Scan(&rowCount); err != nil {
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
	_, err := connection.Conn.Exec(statement)

	if err != nil {
		log.Panicf("Could not execute database statement %s: %s", statement, err)
	}
}

func initIfNeeded() {
	if connection == nil {
		conn, err := db.Initialize(
			"postgres",
			"integration-test-pa55word!",
			"subscriptions",
			"localhost",
			1334)

		if err != nil {
			log.Panicf("Could not connect to database to reset for integration tests: %s", err)
		}

		connection = &conn
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
		MaxElapsedTime:      5 * time.Second,
		Stop:                -1,
		Clock:               backoff.SystemClock,
	})

	if err != nil {
		t.Fatal(err)
	}
}
