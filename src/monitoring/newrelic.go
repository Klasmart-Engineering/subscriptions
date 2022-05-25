package monitoring

import (
	newrelic "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrzap"
)

var newRelicApplication *newrelic.Application

func SetupNewRelic(serviceName string, licenseKey string, enabled, tracerEnabled, spanEventsEnabled, errorCollectorEnabled bool) {
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

	newRelicApplication = &app
	GlobalContext.NewRelic = &app
}
