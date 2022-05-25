package monitoring

import (
	"context"
	newrelic "github.com/newrelic/go-agent/v3/newrelic"
	"go.uber.org/zap"
	"subscriptions/src/config"
)

const (
	traceIDKey    = "trace.id"
	spanIDKey     = "span.id"
	entityGUIDKey = "entity.guid"
	entityNameKey = "entity.name"
	entityTypeKey = "entity.type"
	hostnameKey   = "hostname"
)

var GlobalContext *Context

type Context struct {
	context.Context
	*zap.Logger
	NewRelic *newrelic.Application
}

func SetupGlobalMonitoringContext(ctx context.Context) {
	var l *zap.Logger
	if config.GetConfig().Logging.DevelopmentLogger {
		l, _ = zap.NewDevelopment()
	} else {
		l, _ = zap.NewProduction()
	}

	GlobalContext = NewMonitoringContext(l, ctx)
}

func NewMonitoringContext(logger *zap.Logger, context context.Context) *Context {
	return &Context{
		Context:  context,
		Logger:   logger.With(keyAndValueFromContext(context)...),
		NewRelic: newRelicApplication,
	}
}

func keyAndValueFromContext(ctx context.Context) []zap.Field {
	if txn := newrelic.FromContext(ctx); nil != txn {
		metadata := txn.GetLinkingMetadata()
		return []zap.Field{
			zap.String(traceIDKey, metadata.TraceID),
			zap.String(spanIDKey, metadata.SpanID),
			zap.String(entityGUIDKey, metadata.EntityGUID),
			zap.String(entityNameKey, metadata.EntityName),
			zap.String(entityTypeKey, metadata.EntityType),
			zap.String(hostnameKey, metadata.Hostname),
		}
	}
	return nil
}
