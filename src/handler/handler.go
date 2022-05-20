package handler

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	newrelic "github.com/newrelic/go-agent"
	"net/http"
	conf "subscriptions/config"
	db "subscriptions/database"
)

var dbInstance db.Database
var config conf.Config
var cntxt context.Context

func NewHandler(db db.Database, newRelicApp newrelic.Application, cfg *conf.Config, ctx context.Context) http.Handler {
	router := chi.NewRouter()
	dbInstance = db
	config = *cfg
	cntxt = ctx
	router.MethodNotAllowed(methodNotAllowedHandler)
	router.NotFound(notFoundHandler)
	router.Get(newrelic.WrapHandleFunc(newRelicApp, "/healthcheck", dbHealthcheck))
	router.Get(newrelic.WrapHandleFunc(newRelicApp, "/liveness", applicationLiveness))
	router.Get(newrelic.WrapHandleFunc(newRelicApp, "/subscription-types", getAllSubscriptionTypes))
	router.Get(newrelic.WrapHandleFunc(newRelicApp, "/subscription-actions", getAllSubscriptionActions))
	router.Post(newrelic.WrapHandleFunc(newRelicApp, "/log-action", logAccountAction))
	router.Post(newrelic.WrapHandleFunc(newRelicApp, "/add-product", addProduct))
	router.Post(newrelic.WrapHandleFunc(newRelicApp, "/evaluate-subscriptions", evaluateSubscriptionsUsage))
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
