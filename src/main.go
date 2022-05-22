package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"log"
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

	log.Printf("Starting API server with profile: %s", config.GetProfileName())

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(activeConfig.Server.Port))
	if err != nil {
		log.Fatalf("Error occurred: %s", err.Error())
	}

	database := setupDatabase()
	defer database.Conn.Close()

	newRelic, _ := instrument.GetNewRelic("Subscription Service",
		getLogger(),
		activeConfig.NewRelicConfig.LicenseKey,
		activeConfig.NewRelicConfig.Enabled,
		activeConfig.NewRelicConfig.TracerEnabled,
		activeConfig.NewRelicConfig.SpanEventEnabled,
		activeConfig.NewRelicConfig.ErrorCollectorEnabled)

	httpHandler := handler.NewHandler(database, newRelic.App, ctx)

	server := &http.Server{
		Handler: httpHandler,
	}
	go func() {
		server.Serve(listener)
	}()
	defer Stop(server)
	log.Printf("Started server on %d", activeConfig.Server.Port)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(fmt.Sprint(<-ch))
	log.Println("Stopping API server.")
}

func Stop(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Could not shut down server correctly: %v\n", err)
		os.Exit(1)
	}
}

func getLogger() *logging.ZapLogger {
	var l *zap.Logger
	if config.GetConfig().Logging.DevelopmentLogger {
		l, _ = zap.NewDevelopment()
	} else {
		l, _ = zap.NewProduction()
	}

	return logging.Wrap(l)
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
		log.Fatalf("Could not set up database: %v", err)
	}

	return database
}
