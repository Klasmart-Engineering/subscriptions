package handler

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"subscriptions/src/models"
	"subscriptions/src/monitoring"
)

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
	subscriptionTypes, err := dbInstance.GetSubscriptionTypes(monitoringContext)
	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}
	if err := render.Render(w, r, subscriptionTypes); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func getAllSubscriptionActions(monitoringContext *monitoring.Context, w http.ResponseWriter, r *http.Request) {
	subscriptionActions, err := dbInstance.GetAllSubscriptionActions(monitoringContext)
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
		go logActionWithRecover(monitoring.GlobalContext, action)
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

	var actionResponse = LogAction(monitoringContext, accountAction)

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

	err = dbInstance.AddProductToSubscription(monitoringContext, product)

	if err != nil {
		render.Render(w, r, ErrorRenderer(err))
	} else {
		response := models.ProductResponse{Details: "Successfully added product"}

		if err := render.Render(w, r, &response); err != nil {
			render.Render(w, r, ErrorRenderer(err))
		}
	}
}

func createOrGetSubscription(monitoringContext *monitoring.Context, w http.ResponseWriter, r *http.Request) {
	accountId := chi.URLParam(r, "accountId")
	subId, subscriptionState, err := dbInstance.SubscriptionExists(monitoringContext, accountId)

	if err != nil {
		if subId == uuid.Nil {
			subId, err := dbInstance.CreateSubscription(monitoringContext, accountId)
			if err != nil {
				render.Render(w, r, ErrorRenderer(err))
			}

			response := models.SubscriptionResponse{SubscriptionId: subId.String(), Active: true}
			if err := render.Render(w, r, &response); err != nil {
				render.Render(w, r, ErrorRenderer(err))
			}
		} else {
			render.Render(w, r, ErrorRenderer(err))
		}
	} else {
		var active bool
		if subscriptionState == 1 {
			active = true
		} else {
			active = false
		}

		response := models.SubscriptionResponse{SubscriptionId: subId.String(), Active: active}

		if err := render.Render(w, r, &response); err != nil {
			render.Render(w, r, ErrorRenderer(err))
		}
	}
}

func deactivateSubscription(monitoringContext *monitoring.Context, w http.ResponseWriter, r *http.Request) {
	subscriptionId := chi.URLParam(r, "id")
	var inactiveState = 2 //Inactive
	err := dbInstance.UpdateSubscriptionStatus(monitoringContext, subscriptionId, inactiveState)

	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}

	response := models.GenericResponse{Details: "Subscription deactivated."}
	if err := render.Render(w, r, &response); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func deleteSubscription(monitoringContext *monitoring.Context, w http.ResponseWriter, r *http.Request) {
	subscriptionId := chi.URLParam(r, "id")
	var inactiveState = 3 //Deleted
	err := dbInstance.UpdateSubscriptionStatus(monitoringContext, subscriptionId, inactiveState)

	if err != nil {
		render.Render(w, r, ServerErrorRenderer(err))
		return
	}

	response := models.GenericResponse{Details: "Subscription deleted."}
	if err := render.Render(w, r, &response); err != nil {
		render.Render(w, r, ErrorRenderer(err))
	}
}

func logActionWithRecover(monitoringContext *monitoring.Context, action models.SubscriptionAccountAction) {
	defer func() {
		if r := recover(); r != nil {
			monitoringContext.Error("Something went wrong logging action", zap.Any("error", r))
		}
	}()

	logAction := LogAction(monitoringContext, action)
	monitoringContext.Info(logAction.Details)
}

func LogAction(monitoringContext *monitoring.Context, accountAction models.SubscriptionAccountAction) models.LogResponse {

	dbInstance.LogUserAction(monitoringContext, accountAction)
	interactions, err := dbInstance.CountInteractionsForSubscription(monitoringContext, accountAction)
	if err != nil {
		panic(err)
	}

	threshold, er := dbInstance.GetThresholdForSubscriptionProduct(monitoringContext, accountAction)
	if er != nil {
		panic(er)
	}

	active, err := dbInstance.IsSubscriptionActive(monitoringContext, accountAction.SubscriptionId)

	if err != nil {
		panic(err)
	}

	if !active {
		dbInstance.UpdateChargeableLog(monitoringContext, accountAction)
		return models.LogResponse{Success: false, Details: "BLOCKED. Subscription not active", Count: interactions, Limit: threshold}
	}

	if threshold != 0 && interactions > threshold {
		dbInstance.UpdateChargeableLog(monitoringContext, accountAction)
		return models.LogResponse{Success: false, Details: "BLOCKED", Count: interactions, Limit: threshold}
	}

	return models.LogResponse{Success: true, Details: "WITHIN LIMITS", Count: interactions, Limit: threshold}
}
