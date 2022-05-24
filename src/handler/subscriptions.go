package handler

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strconv"
	"subscriptions/src/models"
	"subscriptions/src/monitoring"
	"time"
)

func evaluateSubscriptionsUsage(monitoringContext *monitoring.Context, w http.ResponseWriter, r *http.Request) {
	subscriptions, err := dbInstance.SubscriptionsToProcess()
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}

	for _, subscriptionToEvaluate := range subscriptions.SubscriptionEvaluations {
		EvaluateSubscription(monitoringContext, subscriptionToEvaluate)
	}

	// do this as part of a transaction
}

func EvaluateSubscription(monitoringContext *monitoring.Context, subscriptionToEvaluate models.SubscriptionEvaluation) {
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

	//TODO revert this back to putting on a topic
	monitoringContext.Info(fmt.Sprint(evaluatedSubscription))
	dbInstance.UpdateLastProcessed(monitoringContext, &subscriptionToEvaluate)
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

func getAllSubscriptionTypes(monitoringContext *monitoring.Context, w http.ResponseWriter, r *http.Request) {
	subscriptionTypes, err := dbInstance.GetSubscriptionTypes()
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	if err := render.Render(w, r, subscriptionTypes); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func getAllSubscriptionActions(monitoringContext *monitoring.Context, w http.ResponseWriter, r *http.Request) {
	subscriptionActions, err := dbInstance.GetAllSubscriptionActions()
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	if err := render.Render(w, r, subscriptionActions); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func logAccountActions(monitoringContext *monitoring.Context, w http.ResponseWriter, r *http.Request) {
	var actionList models.SubscriptionAccountActionList
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	err = json.Unmarshal(bytes, &actionList)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	for _, action := range actionList.Actions {
		go logActionWithRecover(monitoringContext, action)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Account Actions Processing"))
}

func logAccountAction(monitoringContext *monitoring.Context, w http.ResponseWriter, r *http.Request) {
	var accountAction models.SubscriptionAccountAction
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	err = json.Unmarshal(bytes, &accountAction)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	var actionResponse = LogAction(accountAction)

	if err := render.Render(w, r, &actionResponse); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func addProduct(monitoringContext *monitoring.Context, w http.ResponseWriter, r *http.Request) {
	var product models.AddProduct
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	err = json.Unmarshal(bytes, &product)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	err = AddProductToSubscription(product)

	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	} else {
		response := models.ProductResponse{Details: "Successfully added product"}

		if err := render.Render(w, r, &response); err != nil {
			render.Render(w, r, ErrorRenderer(err))
		}
	}
}

func createSubscription(monitoringContext *monitoring.Context, w http.ResponseWriter, r *http.Request) {
	subscription, err := dbInstance.CreateSubscription(monitoringContext)

	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	} else {
		response := models.SubscriptionResponse{SubscriptionId: subscription.String()}

		if err := render.Render(w, r, &response); err != nil {
			render.Render(w, r, ErrorRenderer(err))
		}
	}
}

func deactivateSubscription(monitoringContext *monitoring.Context, w http.ResponseWriter, r *http.Request) {
	subscriptionId := chi.URLParam(r, "id")
	var inactiveState = 2 //Inactive
	err := dbInstance.UpdateSubscriptionStatus(subscriptionId, inactiveState)

	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}

	response := models.GenericResponse{Details: "Subscription deactivated."}
	if err := render.Render(w, r, &response); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func AddProductToSubscription(product models.AddProduct) error {
	err := dbInstance.AddProductToSubscription(product)
	return err
}

func logActionWithRecover(monitoringContext *monitoring.Context, action models.SubscriptionAccountAction) {
	defer func() {
		if r := recover(); r != nil {
			monitoringContext.Error("Something went wrong logging action", zap.Any("error", r))
		}
	}()

	logAction := LogAction(action)
	monitoringContext.Info(logAction.Details)
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
		dbInstance.UpdateChargeableLog(accountAction)
		return models.LogResponse{Success: false, Details: "BLOCKED. Subscription not active", Count: interactions, Limit: threshold}
	}

	if threshold != 0 && interactions > threshold {
		dbInstance.UpdateChargeableLog(accountAction)
		return models.LogResponse{Success: false, Details: "BLOCKED", Count: interactions, Limit: threshold}
	}

	return models.LogResponse{Success: true, Details: "WITHIN LIMITS", Count: interactions, Limit: threshold}
}
