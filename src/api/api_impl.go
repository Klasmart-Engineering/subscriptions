package api

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"strings"
	db "subscriptions/src/database"
	"subscriptions/src/monitoring"
	"subscriptions/src/security"
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
	if result := permissionCheck(ctx, monitoringContext, "create-subscription"); !result {
		return nil
	}

	monitoringContext.Info(fmt.Sprintf("Ok then %+v", request))

	return nil
}

func permissionCheck(ctx echo.Context, monitoringContext *monitoring.Context, permission string) bool {
	bearerToken := strings.Replace(ctx.Request().Header.Get("Authorization"), "Bearer ", "", 1)
	keyMatched, permissionMatched := security.CheckApiKey(monitoringContext, bearerToken, permission)

	if !keyMatched {
		noContentOrLog(monitoringContext, ctx, 401)
		return false
	}

	if !permissionMatched {
		noContentOrLog(monitoringContext, ctx, 403)
		return false
	}

	return true
}
