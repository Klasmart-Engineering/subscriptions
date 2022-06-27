package api

import (
	uuid2 "github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	db "subscriptions/src/database"
	"subscriptions/src/models"
	"subscriptions/src/monitoring"
	"subscriptions/src/services"
	"subscriptions/src/utils"
	"time"
)

type Impl struct{}

var Implementation = &Impl{}

//Your IDE should tell you here if you're not implementing all the endpoints
var _ ServerInterface = (*Impl)(nil)

func (i Impl) GetSubscriptionsSubscriptionIdUsageReportsUsageReportId(ctx echo.Context, monitoringContext *monitoring.Context, apiAuth ApiAuth, subscriptionId string, usageReportId string) error {
	usageReportUUID, err := uuid2.Parse(usageReportId)
	if err != nil {
		noContentOrLog(monitoringContext, ctx, http.StatusBadRequest)
		return nil
	}

	exists, subscription, err := db.GetSubscriptionById(monitoringContext, subscriptionId)
	if err != nil {
		monitoringContext.Error("Unable to check if Subscription exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
		return nil
	}

	if !exists {
		noContentOrLog(monitoringContext, ctx, http.StatusNotFound)
		return nil
	}

	if apiAuth.Jwt == nil || apiAuth.Jwt.AccountId != subscription.AccountId.String() {
		noContentOrLog(monitoringContext, ctx, http.StatusForbidden)
		return nil
	}

	usageReportExists, usageReport, err := db.GetUsageReport(monitoringContext, usageReportUUID)
	if err != nil {
		monitoringContext.Error("Unable to check if Usage Report exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
		return nil
	}

	if !usageReportExists {
		noContentOrLog(monitoringContext, ctx, http.StatusNotFound)
		return nil
	}

	if usageReport.SubscriptionId != subscription.Id {
		noContentOrLog(monitoringContext, ctx, http.StatusForbidden)
		return nil
	}

	usageReportInstances, err := services.CheckUsageReportInstances(monitoringContext, usageReportUUID)
	if err != nil {
		monitoringContext.Error("Unable to get Usage Report instances", zap.Error(err), zap.String("usageReportId", usageReportId))
		noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
		return nil
	}

	var newestCompletedInstance *models.UsageReportInstance
	var pendingInstance *models.UsageReportInstance
	for _, instance := range usageReportInstances {
		instancee := instance
		if (newestCompletedInstance == nil || newestCompletedInstance.RequestedAt.Before(instance.RequestedAt)) && !utils.IsNil(instance.CompletedAt) {
			newestCompletedInstance = &instancee
		}

		if instance.CompletedAt == nil {
			pendingInstance = &instancee
		}
	}

	from := utils.GetMonth(usageReport.Year, usageReport.Month)
	to := utils.ToNextMonth(from)

	state := "not_requested"
	if pendingInstance != nil {
		state = "processing"
	} else if newestCompletedInstance != nil {
		state = "ready"
	}

	if newestCompletedInstance != nil {
		instanceProducts, err := db.GetUsageReportInstanceProducts(monitoringContext, newestCompletedInstance.Id)
		if err != nil {
			monitoringContext.Error("Unable to get Usage Report Instance Products", zap.Error(err), zap.String("usageReportInstanceId", newestCompletedInstance.Id.String()))
			noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
			return nil
		}

		products := UsageReport_Products{}

		for _, product := range instanceProducts {
			products.Set(product.Product, product.Value)
		}

		jsonContentOrLog(monitoringContext, ctx, http.StatusOK, UsageReport{
			Id:                usageReportUUID,
			From:              from.Unix(),
			To:                to.Unix(),
			ReportCompletedAt: utils.Int64Ptr(newestCompletedInstance.CompletedAt.Unix()),
			State:             state,
			Products:          &products,
		})
		return nil
	}

	jsonContentOrLog(monitoringContext, ctx, http.StatusOK, UsageReport{
		Id:                usageReportUUID,
		From:              from.Unix(),
		To:                to.Unix(),
		ReportCompletedAt: nil,
		State:             state,
		Products:          nil,
	})
	return nil
}

func (i Impl) GetSubscriptionsSubscriptionIdUsageReports(ctx echo.Context, monitoringContext *monitoring.Context, apiAuth ApiAuth, subscriptionId string) error {
	exists, subscription, err := db.GetSubscriptionById(monitoringContext, subscriptionId)
	if err != nil {
		monitoringContext.Error("Unable to check if Subscription exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
		return nil
	}

	if !exists {
		noContentOrLog(monitoringContext, ctx, http.StatusNotFound)
		return nil
	}

	if apiAuth.Jwt == nil || apiAuth.Jwt.AccountId != subscription.AccountId.String() {
		noContentOrLog(monitoringContext, ctx, http.StatusForbidden)
		return nil
	}

	usageReports, err := services.GenerateMissingUsageReports(monitoringContext, subscription)
	if err != nil {
		monitoringContext.Error("Unable to get usage reports for Subscription",
			zap.Error(err), zap.String("subscriptionId", subscriptionId))
		noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
		return nil
	}

	response := make([]UsageReports, len(usageReports))
	for i, report := range usageReports {
		from := utils.GetMonth(report.Year, report.Month)
		to := utils.ToNextMonth(from)
		response[i] = UsageReports{Id: report.Id, From: from.Unix(), To: to.Unix()}
	}

	err = ctx.JSON(http.StatusOK, response)
	if err != nil {
		return err
	}

	return nil
}

func (i Impl) PatchSubscriptionsSubscriptionIdUsageReportsUsageReportId(ctx echo.Context, monitoringContext *monitoring.Context, apiAuth ApiAuth, subscriptionId string, usageReportId string) error {
	usageReportUUID, err := uuid2.Parse(usageReportId)
	if err != nil {
		noContentOrLog(monitoringContext, ctx, http.StatusBadRequest)
		return nil
	}

	exists, subscription, err := db.GetSubscriptionById(monitoringContext, subscriptionId)
	if err != nil {
		monitoringContext.Error("Unable to check if Subscription exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
		return nil
	}

	if !exists {
		noContentOrLog(monitoringContext, ctx, http.StatusNotFound)
		return nil
	}

	if apiAuth.Jwt == nil || apiAuth.Jwt.AccountId != subscription.AccountId.String() {
		noContentOrLog(monitoringContext, ctx, http.StatusForbidden)
		return nil
	}

	usageReportExists, usageReport, err := db.GetUsageReport(monitoringContext, usageReportUUID)
	if err != nil {
		monitoringContext.Error("Unable to check if Usage Report exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
		return nil
	}

	if !usageReportExists {
		noContentOrLog(monitoringContext, ctx, http.StatusNotFound)
		return nil
	}

	if usageReport.SubscriptionId != subscription.Id {
		noContentOrLog(monitoringContext, ctx, http.StatusForbidden)
		return nil
	}

	usageReportInstances, err := services.CheckUsageReportInstances(monitoringContext, usageReportUUID)
	if err != nil {
		monitoringContext.Error("Unable to get Usage Report instances", zap.Error(err), zap.String("usageReportId", usageReportId))
		noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
		return nil
	}

	for _, instance := range usageReportInstances {
		if instance.CompletedAt == nil {
			jsonContentOrLog(monitoringContext, ctx, http.StatusOK, UsageReportState{State: "processing"})
			return nil
		}
	}

	err = services.CreateReportInstance(monitoringContext, usageReport)
	if err != nil {
		monitoringContext.Error("Could not create report instance",
			zap.String("subscriptionId", subscriptionId), zap.Int("month", usageReport.Month),
			zap.Int("year", usageReport.Year), zap.Error(err))
		noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
		return nil
	}

	jsonContentOrLog(monitoringContext, ctx, http.StatusOK, UsageReportState{State: "processing"})
	return nil
}

func (Impl) GetHealthcheck(ctx echo.Context, monitoringContext *monitoring.Context) error {
	healthcheck, err := db.Healthcheck()
	if err != nil || !healthcheck {
		err = ctx.JSON(http.StatusInternalServerError, ApplicationStateResponse{
			Up:      false,
			Details: "Could not query the database",
		})
		return err
	}

	err = ctx.JSON(http.StatusOK, ApplicationStateResponse{
		Up:      true,
		Details: "Successfully connected to the database",
	})
	if err != nil {
		return err
	}

	return nil
}

func (Impl) GetLiveness(ctx echo.Context, monitoringContext *monitoring.Context) error {
	err := ctx.JSON(http.StatusOK, ApplicationStateResponse{
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

	err = ctx.JSON(http.StatusOK, response)

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

	err = ctx.JSON(http.StatusOK, response)

	return nil
}

func (Impl) PostSubscriptions(ctx echo.Context, monitoringContext *monitoring.Context, apiAuth ApiAuth, request CreateSubscriptionRequest) error {
	exists, _, err := db.GetSubscriptionByAccountId(monitoringContext, request.AccountId.String())
	if err != nil {
		monitoringContext.Error("Unable to check if Subscription already exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
		return nil
	}

	if exists {
		noContentOrLog(monitoringContext, ctx, http.StatusConflict)
		return nil
	}

	subscription := models.Subscription{
		Id:        uuid2.New(),
		AccountId: request.AccountId,
		State:     models.Active,
		CreatedAt: time.Now(),
	}

	err = db.CreateSubscription(monitoringContext, subscription)
	if err != nil {
		monitoringContext.Error("Unable to create Subscription", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
		return nil
	}

	ctx.Response().Header().Set("Location", "/subscriptions/"+subscription.Id.String())
	ctx.Response().WriteHeader(http.StatusCreated)

	return nil
}

func (i Impl) GetSubscriptionsSubscriptionId(ctx echo.Context, monitoringContext *monitoring.Context, apiAuth ApiAuth, subscriptionId string) error {
	exists, subscription, err := db.GetSubscriptionById(monitoringContext, subscriptionId)
	if err != nil {
		monitoringContext.Error("Unable to check if Subscription exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
		return nil
	}

	if !exists {
		noContentOrLog(monitoringContext, ctx, http.StatusNotFound)
		return nil
	}

	if apiAuth.ApiKey != nil || (apiAuth.Jwt != nil && apiAuth.Jwt.SubscriptionId == subscription.Id.String()) {
		jsonContentOrLog(monitoringContext, ctx, http.StatusOK, Subscription{
			AccountId: subscription.AccountId,
			Id:        subscription.Id,
			State:     subscription.State.String(),
		})
		return nil
	}

	noContentOrLog(monitoringContext, ctx, http.StatusForbidden)
	return nil
}

func (Impl) GetSubscriptions(ctx echo.Context, monitoringContext *monitoring.Context, apiAuth ApiAuth, params GetSubscriptionsParams) error {
	exists, subscription, err := db.GetSubscriptionByAccountId(monitoringContext, params.AccountId)
	if err != nil {
		monitoringContext.Error("Unable to check if Subscription exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
		return nil
	}

	if exists && apiAuth.ApiKey != nil || (apiAuth.Jwt != nil && apiAuth.Jwt.SubscriptionId == subscription.Id.String()) {
		jsonContentOrLog(monitoringContext, ctx, http.StatusOK, []Subscription{
			{
				AccountId: subscription.AccountId,
				Id:        subscription.Id,
				State:     subscription.State.String(),
			},
		})
		return nil
	}

	jsonContentOrLog(monitoringContext, ctx, http.StatusOK, []Subscription{})
	return nil
}

func (i Impl) PatchSubscriptionsSubscriptionId(ctx echo.Context, monitoringContext *monitoring.Context, apiAuth ApiAuth, request PatchSubscriptionRequest, subscriptionId string) error {
	exists, subscription, err := db.GetSubscriptionById(monitoringContext, subscriptionId)
	if err != nil {
		monitoringContext.Error("Unable to check if Subscription exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
		return nil
	}

	if !exists {
		noContentOrLog(monitoringContext, ctx, http.StatusNotFound)
		return nil
	}

	subscriptionState, err := models.SubscriptionStateFromString(request.State)

	if err != nil {
		monitoringContext.Error("Unable to get subscription state", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, http.StatusBadRequest)
		return nil
	}

	if apiAuth.Jwt != nil && apiAuth.Jwt.AccountId == subscription.AccountId.String() {
		err := db.UpdateSubscriptionStatus(monitoringContext, subscriptionId, subscriptionState)

		if err != nil {
			monitoringContext.Error("Unable to get update subscription state", zap.Error(err))
			noContentOrLog(monitoringContext, ctx, http.StatusInternalServerError)
			return nil
		}

		noContentOrLog(monitoringContext, ctx, http.StatusOK)
	}

	noContentOrLog(monitoringContext, ctx, http.StatusForbidden)
	return nil
}
