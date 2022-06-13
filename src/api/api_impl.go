package api

import (
	"fmt"
	"github.com/labstack/echo/v4"
	db "subscriptions/src/database"
	"subscriptions/src/monitoring"
)

type Impl struct{}

var Implementation = &Impl{}

//Your IDE should tell you here if you're not implementing all the endpoints
var _ ServerInterface = (*Impl)(nil)

func (Impl) GetHealthcheck(ctx echo.Context, monitoringContext *monitoring.Context) error {
	healthcheck, err := db.Healthcheck()
	if err != nil || !healthcheck {
		err = ctx.JSON(500, ApplicationStateResponse{
			Up:      false,
			Details: "Could not query the database",
		})
		return err
	}

	err = ctx.JSON(200, ApplicationStateResponse{
		Up:      true,
		Details: "Successfully connected to the database",
	})
	if err != nil {
		return err
	}

	return nil
}

func (Impl) GetLiveness(ctx echo.Context, monitoringContext *monitoring.Context) error {
	err := ctx.JSON(200, ApplicationStateResponse{
		Up:      true,
		Details: "Application Up",
	})
	if err != nil {
		return err
	}

	return nil
}

func (Impl) GetSubscriptionActions(ctx echo.Context, monitoringContext *monitoring.Context) error {
	actions, err := db.GetAllSubscriptionActions(monitoringContext)
	if err != nil {
		return err
	}

	response := make([]SubscriptionAction, len(actions.Actions))
	for i, action := range actions.Actions {
		response[i] = SubscriptionAction{
			Description: action.Description,
			Name:        action.Name,
			Unit:        action.Unit,
		}
	}

	err = ctx.JSON(200, response)

	return nil
}

func (Impl) GetSubscriptionTypes(ctx echo.Context, monitoringContext *monitoring.Context) error {
	types, err := db.GetSubscriptionTypes(monitoringContext)
	if err != nil {
		return err
	}

	response := make([]SubscriptionType, len(types.Subscriptions))
	for i, action := range types.Subscriptions {
		response[i] = SubscriptionType{
			Id:   action.ID,
			Name: action.Name,
		}
	}

	err = ctx.JSON(200, response)

	return nil
}

func (Impl) PostSubscriptions(ctx echo.Context, monitoringContext *monitoring.Context, request CreateSubscriptionRequest) error {
	monitoringContext.Info(fmt.Sprintf("Ok then %+v", request))

	return nil
}
