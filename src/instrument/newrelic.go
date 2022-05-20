package instrument

import (
	"strconv"
	"subscriptions/log"

	newrelic "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrzap"
)

type NewRelic struct {
	serviceName string
	cfg         newrelic.Config
	App         newrelic.Application
}

func GetNewRelic(serviceName string, logger *log.ZapLogger) (*NewRelic, error) {
	cfg := newrelic.NewConfig(serviceName, MustGetEnv("NEW_RELIC_LICENSE_KEY"))
	isEnabled := false
	isDistributedTracerEnabled := false
	isSpanEventsEnabled := false
	isErrorCollectorEnabled := false

	isEnabled, _ = strconv.ParseBool(MustGetEnv("NEW_RELIC_ENABLED"))
	isDistributedTracerEnabled, _ = strconv.ParseBool(MustGetEnv("DISTRIBUTED_TRACER_ENABLED"))
	isSpanEventsEnabled, _ = strconv.ParseBool(MustGetEnv("SPAN_EVENT_ENABLED"))
	isErrorCollectorEnabled, _ = strconv.ParseBool(MustGetEnv("ERROR_COLLECTOR_ENABLED"))

	cfg.Enabled = isEnabled
	cfg.DistributedTracer.Enabled = isDistributedTracerEnabled
	cfg.SpanEvents.Enabled = isSpanEventsEnabled
	cfg.ErrorCollector.Enabled = isErrorCollectorEnabled
	cfg.Logger = nrzap.Transform(logger.Named("newrelic"))

	app, err := newrelic.NewApplication(cfg)

	if err != nil {
		panic("Failed to setup NewRelic: " + err.Error())
	}

	return &NewRelic{
		serviceName: serviceName,
		cfg:         cfg,
		App:         app,
	}, nil
}
