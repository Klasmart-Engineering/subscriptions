package log

import (
	"context"
	newrelic "github.com/newrelic/go-agent"
	"go.uber.org/zap"
)

const (
	traceIDKey    = "trace.id"
	spanIDKey     = "span.id"
	entityGUIDKey = "entity.guid"
	entityNameKey = "entity.name"
	entityTypeKey = "entity.type"
	hostnameKey   = "hostname"
)

var GlobalContext *SubscriptionsContext

type SubscriptionsContext struct {
	context.Context
	*zap.Logger
}

func NewSubscriptionsContext(logger *zap.Logger, context context.Context) *SubscriptionsContext {
	return &SubscriptionsContext{
		Context: context,
		Logger:  logger.With(keyAndValueFromContext(context)...),
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
