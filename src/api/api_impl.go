package api

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	uuid2 "github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"strings"
	"subscriptions/src/aws"
	db "subscriptions/src/database"
	"subscriptions/src/models"
	"subscriptions/src/monitoring"
	"subscriptions/src/utils"
	"time"
)

type Impl struct{}

var Implementation = &Impl{}

//Your IDE should tell you here if you're not implementing all the endpoints
var _ ServerInterface = (*Impl)(nil)

func (i Impl) GetSubscriptionsSubscriptionIdUsageReportsUsageReportId(ctx echo.Context, monitoringContext *monitoring.Context, apiAuth ApiAuth, subscriptionId string, usageReportId string) error {
	exists, subscription, err := db.GetSubscriptionById(monitoringContext, subscriptionId)
	if err != nil {
		monitoringContext.Error("Unable to check if Subscription exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, 500)
		return nil
	}

	if !exists {
		noContentOrLog(monitoringContext, ctx, 404)
		return nil
	}

	if apiAuth.Jwt == nil || apiAuth.Jwt.AccountId != subscription.AccountId.String() {
		noContentOrLog(monitoringContext, ctx, 403)
		return nil
	}

	//TEMP until S3 & Athena implementation
	report := UsageReport{Id: uuid2.New(), From: "1640995200", To: "1640995200", ReportCompletedAt: "1640991100", State: "ready", Products: Product{ProductName: "Content API"}}

	err = ctx.JSON(200, report)
	if err != nil {
		return err
	}

	return nil
}

func (i Impl) GetSubscriptionsSubscriptionIdUsageReports(ctx echo.Context, monitoringContext *monitoring.Context, apiAuth ApiAuth, subscriptionId string) error {
	exists, subscription, err := db.GetSubscriptionById(monitoringContext, subscriptionId)
	if err != nil {
		monitoringContext.Error("Unable to check if Subscription exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, 500)
		return nil
	}

	if !exists {
		noContentOrLog(monitoringContext, ctx, 404)
		return nil
	}

	if apiAuth.Jwt == nil || apiAuth.Jwt.AccountId != subscription.AccountId.String() {
		noContentOrLog(monitoringContext, ctx, 403)
		return nil
	}

	//TEMP until S3 & Athena implementation
	response := make([]UsageReports, 3)
	for i := 0; i < 3; i++ {
		response[i] = UsageReports{Id: uuid2.New(), From: "1640995200", To: "1643673599"}
	}

	err = ctx.JSON(200, response)
	if err != nil {
		return err
	}

	return nil
}

func (i Impl) PatchSubscriptionsSubscriptionIdUsageReportsUsageReportId(ctx echo.Context, monitoringContext *monitoring.Context, apiAuth ApiAuth, subscriptionId string, usageReportId string) error {
	exists, subscription, err := db.GetSubscriptionById(monitoringContext, subscriptionId)
	if err != nil {
		monitoringContext.Error("Unable to check if Subscription exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, 500)
		return nil
	}

	if !exists {
		noContentOrLog(monitoringContext, ctx, 404)
		return nil
	}

	if apiAuth.Jwt == nil || apiAuth.Jwt.AccountId != subscription.AccountId.String() {
		noContentOrLog(monitoringContext, ctx, 403)
		return nil
	}

	//TEMP until S3 & Athena implementation
	state := UsageReportState{State: "processing"}

	err = ctx.JSON(200, state)
	if err != nil {
		return err
	}

	return nil
}

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

func (Impl) PostSubscriptions(ctx echo.Context, monitoringContext *monitoring.Context, apiAuth ApiAuth, request CreateSubscriptionRequest) error {
	exists, _, err := db.GetSubscriptionByAccountId(monitoringContext, request.AccountId.String())
	if err != nil {
		monitoringContext.Error("Unable to check if Subscription already exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, 500)
		return nil
	}

	if exists {
		noContentOrLog(monitoringContext, ctx, 409)
		return nil
	}

	subscription := models.Subscription{
		Id:        uuid2.New(),
		AccountId: request.AccountId,
		State:     models.Active,
	}

	err = db.CreateSubscription(monitoringContext, subscription)
	if err != nil {
		monitoringContext.Error("Unable to create Subscription", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, 500)
		return nil
	}

	ctx.Response().Header().Set("Location", "/subscriptions/"+subscription.Id.String())
	ctx.Response().WriteHeader(201)

	return nil
}

func (i Impl) GetSubscriptionsSubscriptionId(ctx echo.Context, monitoringContext *monitoring.Context, apiAuth ApiAuth, subscriptionId string) error {
	exists, subscription, err := db.GetSubscriptionById(monitoringContext, subscriptionId)
	if err != nil {
		monitoringContext.Error("Unable to check if Subscription exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, 500)
		return nil
	}

	if !exists {
		noContentOrLog(monitoringContext, ctx, 404)
		return nil
	}

	if apiAuth.ApiKey != nil || (apiAuth.Jwt != nil && apiAuth.Jwt.SubscriptionId == subscription.Id.String()) {
		jsonContentOrLog(monitoringContext, ctx, 200, Subscription{
			AccountId: subscription.AccountId,
			Id:        subscription.Id,
			State:     subscription.State.String(),
		})
		return nil
	}

	noContentOrLog(monitoringContext, ctx, 403)
	return nil
}

func (Impl) GetSubscriptions(ctx echo.Context, monitoringContext *monitoring.Context, apiAuth ApiAuth, params GetSubscriptionsParams) error {
	exists, subscription, err := db.GetSubscriptionByAccountId(monitoringContext, params.AccountId)
	if err != nil {
		monitoringContext.Error("Unable to check if Subscription exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, 500)
		return nil
	}

	if exists && apiAuth.ApiKey != nil || (apiAuth.Jwt != nil && apiAuth.Jwt.SubscriptionId == subscription.Id.String()) {
		jsonContentOrLog(monitoringContext, ctx, 200, []Subscription{
			{
				AccountId: subscription.AccountId,
				Id:        subscription.Id,
				State:     subscription.State.String(),
			},
		})
		return nil
	}

	jsonContentOrLog(monitoringContext, ctx, 200, []Subscription{})
	return nil
}

func (i Impl) PatchSubscriptionsSubscriptionId(ctx echo.Context, monitoringContext *monitoring.Context, apiAuth ApiAuth, request PatchSubscriptionRequest, subscriptionId string) error {
	exists, subscription, err := db.GetSubscriptionById(monitoringContext, subscriptionId)
	if err != nil {
		monitoringContext.Error("Unable to check if Subscription exists", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, 500)
		return nil
	}

	if !exists {
		noContentOrLog(monitoringContext, ctx, 404)
		return nil
	}

	subscriptionState, err := models.SubscriptionStateFromString(request.State)

	if err != nil {
		monitoringContext.Error("Unable to get subscription state", zap.Error(err))
		noContentOrLog(monitoringContext, ctx, 400)
		return nil
	}

	if apiAuth.Jwt != nil && apiAuth.Jwt.AccountId == subscription.AccountId.String() {
		err := db.UpdateSubscriptionStatus(monitoringContext, subscriptionId, subscriptionState)

		if err != nil {
			monitoringContext.Error("Unable to get update subscription state", zap.Error(err))
			noContentOrLog(monitoringContext, ctx, 500)
			return nil
		}

		noContentOrLog(monitoringContext, ctx, 200)
	}

	noContentOrLog(monitoringContext, ctx, 403)
	return nil
}

func (i Impl) GetTestAthena(ctx echo.Context, monitoringContext *monitoring.Context, params GetTestAthenaParams) error {
	if _, err := uuid2.Parse(*params.SubscriptionId); err != nil {
		monitoringContext.Error("Could not parse subscription id query param", zap.Error(err))
		ctx.NoContent(400)
		return nil
	}

	tableName := fmt.Sprintf("usage_report_%s_%s",
		strings.ReplaceAll(*params.SubscriptionId, "-", "_"), time.Now().Format("2006_01"))

	createTableDDL := fmt.Sprintf(`CREATE EXTERNAL TABLE IF NOT EXISTS %s (
		id STRING,
		occurred_at BIGINT,
		product STRING,
		method STRING,
		path STRING,
		android_id STRING,
		subscription_id STRING
	) ROW FORMAT SERDE 'org.openx.data.jsonserde.JsonSerDe' 
	LOCATION 's3://subscriptions-uk-apifactory-api-usage-firehose/%s/%s/'`,
		tableName, *params.SubscriptionId, time.Now().Format("2006/01"))

	ddlResponse, err := aws.AthenaClient.StartQueryExecution(monitoringContext, &athena.StartQueryExecutionInput{
		QueryString: &createTableDDL,
		QueryExecutionContext: &types.QueryExecutionContext{
			Database: utils.StringPtr("subscriptions_api_usage"),
		},
		WorkGroup: utils.StringPtr("subscriptions-uk-apifactory-subscriptions-athena"),
		ResultConfiguration: &types.ResultConfiguration{
			OutputLocation: utils.StringPtr("s3://subscriptions-uk-apifactory-subscriptions-athena/"),
		},
	})
	if err != nil {
		monitoringContext.Error("Something went wrong creating table", zap.Error(err))
		ctx.NoContent(500)
		return nil
	}

	monitoringContext.Info("Finished create table DDL: " + *ddlResponse.QueryExecutionId)

	ddlResults, err := aws.AthenaClient.GetQueryResults(monitoringContext, &athena.GetQueryResultsInput{
		QueryExecutionId: ddlResponse.QueryExecutionId,
	})

	if err != nil {
		monitoringContext.Error("Something went wrong getting create table results", zap.Error(err))
		ctx.NoContent(500)
		return nil
	}

	monitoringContext.Info(fmt.Sprintf("ddl results %+v", ddlResults))

	monthlyUsageQuery := fmt.Sprintf("SELECT COUNT(1) FROM %s", tableName)

	queryResponse, err := aws.AthenaClient.StartQueryExecution(monitoringContext, &athena.StartQueryExecutionInput{
		QueryString: &monthlyUsageQuery,
		QueryExecutionContext: &types.QueryExecutionContext{
			Database: utils.StringPtr("subscriptions_api_usage"),
		},
		WorkGroup: utils.StringPtr("subscriptions-uk-apifactory-subscriptions-athena"),
		ResultConfiguration: &types.ResultConfiguration{
			OutputLocation: utils.StringPtr("s3://subscriptions-uk-apifactory-subscriptions-athena/"),
		},
	})
	if err != nil {
		monitoringContext.Error("Something went wrong creating table", zap.Error(err))
		ctx.NoContent(500)
		return nil
	}

	monitoringContext.Info("Finished query: " + *queryResponse.QueryExecutionId)

	queryResults, err := aws.AthenaClient.GetQueryResults(monitoringContext, &athena.GetQueryResultsInput{
		QueryExecutionId: queryResponse.QueryExecutionId,
	})

	if err != nil {
		monitoringContext.Error("Something went wrong getting query results", zap.Error(err))
		ctx.NoContent(500)
		return nil
	}

	monitoringContext.Info(fmt.Sprintf("query results %+v", queryResults))

	ctx.NoContent(200)
	return nil
}
