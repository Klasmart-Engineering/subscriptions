package main

import (
	"context"
	"go.uber.org/zap"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"subscriptions/src/config"
	db "subscriptions/src/database"
	"subscriptions/src/handler"
	"subscriptions/src/monitoring"
	"subscriptions/src/utils"
	"syscall"
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

	monitoringContext := getMonitoringContext(ctx)
	monitoring.GlobalContext = monitoringContext

	monitoringContext.Info("Starting Server",
		zap.String("profile", config.GetProfileName()),
		zap.Int("port", activeConfig.Server.Port))

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(activeConfig.Server.Port))
	if err != nil {
		monitoringContext.Fatal("Unable to start Server: %s", zap.Error(err))
	}

	database := setupDatabase()
	defer database.Conn.Close()

	monitoring.SetupNewRelic(activeConfig.NewRelicConfig.EntityName,
		activeConfig.NewRelicConfig.LicenseKey,
		activeConfig.NewRelicConfig.Enabled,
		activeConfig.NewRelicConfig.TracerEnabled,
		activeConfig.NewRelicConfig.SpanEventEnabled,
		activeConfig.NewRelicConfig.ErrorCollectorEnabled)

	httpHandler := handler.NewHandler(database, monitoringContext)

	server := &http.Server{
		Handler: httpHandler,
	}
	go func() {
		server.Serve(listener)
	}()
	defer Stop(server)
	monitoringContext.Info("Started Server")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	monitoringContext.Info("Stopping Server")
}

func Stop(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		monitoring.GlobalContext.Error("Could not shut down server correctly", zap.Error(err))
		os.Exit(1)
	}
}

func getMonitoringContext(ctx context.Context) *monitoring.Context {
	var l *zap.Logger
	if config.GetConfig().Logging.DevelopmentLogger {
		l, _ = zap.NewDevelopment()
	} else {
		l, _ = zap.NewProduction()
	}

	return monitoring.NewMonitoringContext(l, ctx)
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
