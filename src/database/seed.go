package db

import (
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"os"
	"subscriptions/src/config"
	"subscriptions/src/monitoring"
)

func seedDatabase() {
	if !config.GetConfig().Database.Seed {
		return
	}

	var exists bool
	err := dbConnection.Get(&exists, "SELECT EXISTS (SELECT FROM information_schema.tables where table_schema = 'public' AND table_name = 'seeded')")
	if err != nil {
		monitoring.GlobalContext.Fatal("Could not check if seed marker table exists", zap.Error(err))
		return
	}

	if exists {
		monitoring.GlobalContext.Info("Database is already seeded")
		return
	}

	monitoring.GlobalContext.Info("Starting database seed")

	_, err = dbConnection.Exec(readFile("./database/seed.sql"))
	if err != nil {
		monitoring.GlobalContext.Fatal("Could not run seed script", zap.Error(err))
		return
	}

	_, err = dbConnection.Exec("CREATE TABLE seeded (seeded boolean)")
	if err != nil {
		monitoring.GlobalContext.Fatal("Could not create seed marker table")
		return
	}

	monitoring.GlobalContext.Info("Finished database seed")
}

func readFile(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		monitoring.GlobalContext.Fatal("Could not read: "+file, zap.Error(err))
	}
	return string(content)
}
