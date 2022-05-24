package monitoring

import (
	newrelic "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrzap"
)

type NewRelic struct {
	serviceName string
	cfg         newrelic.Config
	App         newrelic.Application
}

func GetNewRelic(serviceName string, licenseKey string, enabled, tracerEnabled, spanEventsEnabled, errorCollectorEnabled bool) (*NewRelic, error) {
	cfg := newrelic.NewConfig(serviceName, licenseKey)
	cfg.Enabled = enabled
	cfg.DistributedTracer.Enabled = tracerEnabled
	cfg.SpanEvents.Enabled = spanEventsEnabled
	cfg.ErrorCollector.Enabled = errorCollectorEnabled
	cfg.Logger = nrzap.Transform(GlobalContext.Named("newrelic"))

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
