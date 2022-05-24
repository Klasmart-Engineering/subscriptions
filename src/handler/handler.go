package handler

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	newrelic "github.com/newrelic/go-agent"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	db "subscriptions/src/database"
	"subscriptions/src/monitoring"
)

var dbInstance db.Database

func NewHandler(db db.Database, newRelicApp newrelic.Application, ctx *monitoring.Context) http.Handler {
	router := chi.NewRouter()
	dbInstance = db
	router.Use(recovery)
	router.MethodNotAllowed(methodNotAllowedHandler)
	router.NotFound(notFoundHandler)
	router.Get("/healthcheck", dbHealthcheck)
	router.Get("/liveness", applicationLiveness)
	router.Handle("/metrics", promhttp.Handler())
	router.Get(wrap(newRelicApp, ctx, "/subscription-types", getAllSubscriptionTypes))
	router.Get(wrap(newRelicApp, ctx, "/subscription-actions", getAllSubscriptionActions))
	router.Post(wrap(newRelicApp, ctx, "/log-action", logAccountAction))
	router.Post(wrap(newRelicApp, ctx, "/log-actions", logAccountActions))
	router.Post(wrap(newRelicApp, ctx, "/add-product", addProduct))
	router.Post(wrap(newRelicApp, ctx, "/create-subscription", createSubscription))
	router.Post(wrap(newRelicApp, ctx, "/deactivate/{id}", deactivateSubscription))
	router.Post(wrap(newRelicApp, ctx, "/evaluate-subscriptions", evaluateSubscriptionsUsage))

	return router
}

func wrap(newRelicApp newrelic.Application, ctx *monitoring.Context, pattern string, handler func(*monitoring.Context, http.ResponseWriter, *http.Request)) (string, func(http.ResponseWriter, *http.Request)) {
	return newrelic.WrapHandleFunc(newRelicApp, pattern, func(w http.ResponseWriter, r *http.Request) {
		var monitoringContext = monitoring.NewMonitoringContext(ctx.Logger, r.Context())
		monitoringContext.Info("Request started",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()))
		defer monitoringContext.Info("Request finished",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()))
		handler(monitoringContext, w, r)
	})
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(405)
	render.Render(w, r, ErrMethodNotAllowed)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(400)
	render.Render(w, r, ErrNotFound)
}

func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			err := recover()
			if err != nil {
				monitoringContext := monitoring.NewMonitoringContext(monitoring.GlobalContext.Logger, r.Context())
				monitoringContext.Error("Panic caught by recovery handler",
					zap.String("method", r.Method),
					zap.String("requestId", r.RequestURI),
					zap.Any("error", err))

				jsonBody, _ := json.Marshal(map[string]string{
					"error": "There was an internal server error",
				})

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(jsonBody)
			}

		}()

		next.ServeHTTP(w, r)

	})
}
