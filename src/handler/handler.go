package handler

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	newrelic "github.com/newrelic/go-agent"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	db "subscriptions/src/database"
)

var dbInstance db.Database
var cntxt context.Context

func NewHandler(db db.Database, newRelicApp newrelic.Application, ctx context.Context) http.Handler {
	router := chi.NewRouter()
	dbInstance = db
	cntxt = ctx
	router.Use(recovery)
	router.MethodNotAllowed(methodNotAllowedHandler)
	router.NotFound(notFoundHandler)
	router.Get(newrelic.WrapHandleFunc(newRelicApp, "/healthcheck", dbHealthcheck))
	router.Get(newrelic.WrapHandleFunc(newRelicApp, "/liveness", applicationLiveness))
	router.Get(newrelic.WrapHandleFunc(newRelicApp, "/subscription-types", getAllSubscriptionTypes))
	router.Get(newrelic.WrapHandleFunc(newRelicApp, "/subscription-actions", getAllSubscriptionActions))
	router.Post(newrelic.WrapHandleFunc(newRelicApp, "/log-action", logAccountAction))
	router.Post(newrelic.WrapHandleFunc(newRelicApp, "/log-actions", logAccountActions))
	router.Post(newrelic.WrapHandleFunc(newRelicApp, "/add-product", addProduct))
	router.Post(newrelic.WrapHandleFunc(newRelicApp, "/create-subscription", createSubscription))
	router.Post(newrelic.WrapHandleFunc(newRelicApp, "/evaluate-subscriptions", evaluateSubscriptionsUsage))

	router.Handle("/metrics", promhttp.Handler())

	return router
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
				log.Printf("Panic caught by recovery handler on %s request to %s: %s\n", r.Method, r.RequestURI, err)

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
