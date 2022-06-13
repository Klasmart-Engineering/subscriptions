package handler

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/google/uuid"
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
