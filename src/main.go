package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"subscriptions/src/api"
	"subscriptions/src/aws"
	"subscriptions/src/config"
	"subscriptions/src/cron"
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

	go setupDatabase()
	defer db.Close()

	cron.StartCronJobs()
	aws.SetupAWS()

	monitoring.GlobalContext.Info("Starting Server",
		zap.String("profile", config.GetProfileName()),
		zap.Int("port", activeConfig.Server.Port))

	e := echo.New()
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{StackSize: 1 << 10, LogLevel: log.ERROR}))
	api.RegisterHandlers(e, api.Implementation)

	if config.GetConfig().Testing {
		e.Router().Add("POST", "/cron", cron.ForceCronJob)
	}

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

func setupDatabase() {
	for {
		activeConfig := config.GetConfig()

		err := db.Initialize(
			activeConfig.Database.User,
			activeConfig.Database.Password,
			activeConfig.Database.DatabaseName,
			activeConfig.Database.Host,
			activeConfig.Database.Port)

		if err != nil {
			monitoring.GlobalContext.Error("Could not set up database", zap.Error(err))
			time.Sleep(1 * time.Minute)
			continue
		}

		return
	}
}
