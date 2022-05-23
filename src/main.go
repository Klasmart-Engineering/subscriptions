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
	"subscriptions/src/instrument"
	logging "subscriptions/src/log"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startServer(ctx)
}

func startServer(ctx context.Context) {
	config.LoadProfile(instrument.MustGetEnvOrFlag("profile"))
	activeConfig := config.GetConfig()

	subscriptionsContext := getSubscriptionsContext(ctx)
	logging.GlobalContext = subscriptionsContext

	subscriptionsContext.Info("Starting Server",
		zap.String("profile", config.GetProfileName()),
		zap.Int("port", activeConfig.Server.Port))

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(activeConfig.Server.Port))
	if err != nil {
		subscriptionsContext.Fatal("Unable to start Server: %s", zap.Error(err))
	}

	database := setupDatabase()
	defer database.Conn.Close()

	newRelic, _ := instrument.GetNewRelic("Subscription Service",
		activeConfig.NewRelicConfig.LicenseKey,
		activeConfig.NewRelicConfig.Enabled,
		activeConfig.NewRelicConfig.TracerEnabled,
		activeConfig.NewRelicConfig.SpanEventEnabled,
		activeConfig.NewRelicConfig.ErrorCollectorEnabled)

	httpHandler := handler.NewHandler(database, newRelic.App, subscriptionsContext)

	server := &http.Server{
		Handler: httpHandler,
	}
	go func() {
		server.Serve(listener)
	}()
	defer Stop(server)
	subscriptionsContext.Info("Started Server")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	subscriptionsContext.Info("Stopping Server")
}

func Stop(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logging.GlobalContext.Error("Could not shut down server correctly", zap.Error(err))
		os.Exit(1)
	}
}

func getSubscriptionsContext(ctx context.Context) *logging.SubscriptionsContext {
	var l *zap.Logger
	if config.GetConfig().Logging.DevelopmentLogger {
		l, _ = zap.NewDevelopment()
	} else {
		l, _ = zap.NewProduction()
	}

	return logging.NewSubscriptionsContext(l, ctx)
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
		logging.GlobalContext.Fatal("Could not set up database", zap.Error(err))
	}

	return database
}
