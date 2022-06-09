package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"subscriptions/src/api"
	"subscriptions/src/config"
	db "subscriptions/src/database"
	"subscriptions/src/monitoring"
	"subscriptions/src/utils"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startServer(ctx)
}

func startServer(ctx context.Context) {
	config.LoadProfile(utils.MustGetEnvOrFlag("profile"))
	activeConfig := config.GetConfig()

	monitoring.SetupGlobalMonitoringContext(ctx)
	monitoring.SetupNewRelic(activeConfig.NewRelicConfig.EntityName,
		activeConfig.NewRelicConfig.LicenseKey,
		activeConfig.NewRelicConfig.Enabled,
		activeConfig.NewRelicConfig.TracerEnabled)

	database := setupDatabase()
	defer database.Conn.Close()

	monitoring.GlobalContext.Info("Starting Server",
		zap.String("profile", config.GetProfileName()),
		zap.Int("port", activeConfig.Server.Port))

	e := echo.New()
	api.RegisterHandlers(e, api.Implementation)

	go func() {
		if err := e.Start(":" + strconv.Itoa(activeConfig.Server.Port)); err != nil && err != http.ErrServerClosed {
			monitoring.GlobalContext.Fatal("Shutting down Server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func setupDatabase() db.Database {
	activeConfig := config.GetConfig()

	database, err := db.Initialize(
		activeConfig.Database.User,
		activeConfig.Database.Password,
		activeConfig.Database.DatabaseName,
		activeConfig.Database.Host,
		activeConfig.Database.Port)

	if err != nil {
		monitoring.GlobalContext.Fatal("Could not set up database", zap.Error(err))
	}

	return database
}
