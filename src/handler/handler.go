package handler

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.uber.org/zap"
	"net/http"
	db "subscriptions/src/database"
	"subscriptions/src/monitoring"
)

var dbInstance db.Database

func NewHandler(db db.Database, ctx *monitoring.Context) http.Handler {
	router := chi.NewRouter()
	dbInstance = db

	//NOTE: Please add Documentation generation information in generate_docs.go
	router.Use(recovery)
	router.MethodNotAllowed(methodNotAllowedHandler)
	router.NotFound(notFoundHandler)
	router.Get("/healthcheck", dbHealthcheck)
	router.Get("/liveness", applicationLiveness)
	router.Get(wrap(ctx, "/subscription-types", getAllSubscriptionTypes))
	router.Get(wrap(ctx, "/subscription-actions", getAllSubscriptionActions))
	router.Post(wrap(ctx, "/log-action", logAccountAction))
	router.Post(wrap(ctx, "/log-actions", logAccountActions))
	router.Post(wrap(ctx, "/add-product", addProduct))
	router.Get(wrap(ctx, "/subscription/{accountId}", createOrGetSubscription))
	router.Post(wrap(ctx, "/deactivate/{id}", deactivateSubscription))
	router.Post(wrap(ctx, "/delete/{id}", deleteSubscription))

	return router
}

func wrap(ctx *monitoring.Context, pattern string, handler func(*monitoring.Context, http.ResponseWriter, *http.Request)) (string, func(http.ResponseWriter, *http.Request)) {
	return newrelic.WrapHandleFunc(monitoring.GlobalContext.NewRelic, pattern, func(w http.ResponseWriter, r *http.Request) {
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
	w.WriteHeader(404)
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
