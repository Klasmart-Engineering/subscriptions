package monitoring

import (
	"github.com/newrelic/go-agent/v3/integrations/nrzap"
	"github.com/newrelic/go-agent/v3/newrelic"
)

var newRelicApplication *newrelic.Application

func SetupNewRelic(serviceName string, licenseKey string, enabled, tracerEnabled bool) {
	app, err := newrelic.NewApplication(newrelic.ConfigAppName(serviceName),
		newrelic.ConfigEnabled(enabled),
		newrelic.ConfigDistributedTracerEnabled(tracerEnabled),
		newrelic.ConfigLicense(licenseKey),
		newrelic.ConfigLogger(nrzap.Transform(GlobalContext.Named("newrelic"))))
	//newrelic.ConfigDebugLogger(os.Stdout))

	if err != nil {
		panic("Failed to setup NewRelic: " + err.Error())
	}

	newRelicApplication = app
	GlobalContext.NewRelic = app
}
