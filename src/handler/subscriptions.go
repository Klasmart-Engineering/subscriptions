package handler

import (
	"encoding/json"
	"github.com/go-chi/render"
	"net/http"
	"strconv"
	"subscriptions.demo/models"
	"time"
)

func evaluateSubscriptionsUsage(w http.ResponseWriter, r *http.Request) {
	subscriptions, err := dbInstance.SubscriptionsToProcess()
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}

	for _, subscriptionToEvaluate := range subscriptions.SubscriptionEvaluations {
		EvaluateSubscription(subscriptionToEvaluate)
	}

	// look in log for each since last_processed_time ands tally that up
	// also use NOW on that query. Return that and also set that on the JSON response
	// send message to Kafka topic and then update last_processed.
	// do this as part of a transaction
}

func EvaluateSubscription(subscriptionToEvaluate models.SubscriptionEvaluation) {
	productToProductUsage, err := dbInstance.UsageOfSubscription(subscriptionToEvaluate)

	if err != nil {
		panic(err)
	}

	now := time.Now()
	var prods []models.EvaluatedSubscriptionProduct
	for product, usage := range productToProductUsage {
		prods = append(prods, models.EvaluatedSubscriptionProduct{Name: product.Name, Type: product.Type, UsageAmount: usage})
	}
	var evaluatedSubscription = models.EvaluatedSubscription{SubscriptionId: subscriptionToEvaluate.ID, Products: prods, DateFromEpoch: subscriptionToEvaluate.LastProcessedTime, DateToEpoch: strconv.FormatInt(now.Unix(), 10)}

	PublishSubscriptionUsage(evaluatedSubscription, cntxt)
	dbInstance.UpdateLastProcessed(&subscriptionToEvaluate)
}

func dbHealthcheck(w http.ResponseWriter, r *http.Request) {
	up, err := dbInstance.Healthcheck()

	if err != nil || !up {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}

	health := models.Healthcheck{Up: true, Details: "Successfully connected to the database"}
	if err := render.Render(w, r, &health); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func applicationLiveness(w http.ResponseWriter, r *http.Request) {

	health := models.Healthcheck{Up: true, Details: "Application up"}
	if err := render.Render(w, r, &health); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func getAllSubscriptionTypes(w http.ResponseWriter, r *http.Request) {
	subscriptionTypes, err := dbInstance.GetSubscriptionTypes()
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	if err := render.Render(w, r, subscriptionTypes); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func getAllSubscriptionActions(w http.ResponseWriter, r *http.Request) {
	subscriptionActions, err := dbInstance.GetAllSubscriptionActions()
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	if err := render.Render(w, r, subscriptionActions); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func logAccountAction(w http.ResponseWriter, r *http.Request) {
	var accountAction models.SubscriptionAccountAction
	json.NewDecoder(r.Body).Decode(&accountAction)

	var actionResponse = LogAction(accountAction)

	if err := render.Render(w, r, &actionResponse); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func LogAction(accountAction models.SubscriptionAccountAction) models.LogResponse {

	dbInstance.LogUserAction(accountAction)
	interactions, err := dbInstance.CountInteractionsForSubscription(accountAction)
	if err != nil {
		panic(err)
	}

	threshold, er := dbInstance.GetThresholdForSubscriptionProduct(accountAction)
	if er != nil {
		panic(er)
	}

	active, err := dbInstance.IsSubscriptionActive(accountAction.SubscriptionId)

	if err != nil {
		panic(err)
	}

	if !active {
		return models.LogResponse{Success: false, Details: "BLOCKED. Subscription not active", Count: interactions, Limit: threshold}
	}

	if threshold != 0 && interactions > threshold {
		return models.LogResponse{Success: false, Details: "BLOCKED", Count: interactions, Limit: threshold}
	}

	return models.LogResponse{Success: true, Details: "WITHIN LIMITS", Count: interactions, Limit: threshold}
}
