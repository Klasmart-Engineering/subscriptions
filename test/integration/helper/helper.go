package helper

import (
	"fmt"
	"log"
	"os"
	db "subscriptions/src/database"
)

var connection *db.Database
var dropStatements = readFile("../../database/drop-all-tables.sql")
var initStatements = readFile("../../database/init.sql")

func readFile(file string) string {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println("working directory: " + path)

	content, err := os.ReadFile(file)
	if err != nil {
		log.Panicf("Could not read %s", file)
	}
	return string(content)
}

func ResetDatabase() {
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

	execOrPanic(dropStatements)
	execOrPanic(initStatements)
}

func RunTestSql(fileName string) {
	execOrPanic(readFile("./test-sql/" + fileName))
}

func execOrPanic(statement string) {
	_, err := connection.Conn.Exec(statement)

	if err != nil {
		log.Panicf("Could not execute database statement %s: %s", statement, err)
	}
}
