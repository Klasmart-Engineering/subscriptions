package services

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/cenkalti/backoff/v4"
	uuid2 "github.com/google/uuid"
	"go.uber.org/zap"
	"strings"
	"subscriptions/src/aws"
	"subscriptions/src/config"
	db "subscriptions/src/database"
	"subscriptions/src/models"
	"subscriptions/src/monitoring"
	"subscriptions/src/utils"
	"time"
)

type missingMonth struct {
	year  int
	month int
}

func CreateReportInstance(monitoringContext *monitoring.Context, usageReport models.UsageReport) error {
	err := setupTable(monitoringContext, usageReport)
	if err != nil {
		return err
	}

	reportInstance := models.UsageReportInstance{
		Id:            uuid2.New(),
		UsageReportId: usageReport.Id,
		RequestedAt:   time.Now(),
		AthenaQueryId: "",
		CompletedAt:   nil,
	}

	queryId, err := createInstanceQuery(monitoringContext, usageReport)
	if err != nil {
		return err
	}

	reportInstance.AthenaQueryId = queryId
	err = db.InsertUsageReportInstance(monitoringContext, reportInstance)

	return err
}

func GenerateMissingUsageReports(monitoringContext *monitoring.Context, subscription models.Subscription) ([]models.UsageReport, error) {
	currentUsageReports, err := db.GetUsageReportsForSubscription(monitoringContext, subscription.Id)
	if err != nil {
		return nil, err
	}

	missingMonths := getMissingMonths(currentUsageReports, subscription)

	for _, month := range missingMonths {
		usageReport := models.UsageReport{
			Id:             uuid2.New(),
			SubscriptionId: subscription.Id,
			Year:           month.year,
			Month:          month.month,
		}

		err := db.InsertUsageReport(monitoringContext, usageReport)
		if err != nil {
			return nil, err
		}

		currentUsageReports = append(currentUsageReports, usageReport)
	}

	return currentUsageReports, nil
}

func CheckUsageReportInstances(monitoringContext *monitoring.Context, usageReportId uuid2.UUID) ([]models.UsageReportInstance, error) {
	usageReportInstances, err := db.GetUsageReportInstances(monitoringContext, usageReportId)
	if err != nil {
		return nil, err
	}

	for i, instance := range usageReportInstances {
		if instance.CompletedAt == nil {
			execution, err := aws.AthenaClient.GetQueryExecution(monitoringContext, &athena.GetQueryExecutionInput{
				QueryExecutionId: &instance.AthenaQueryId,
			})
			if err != nil {
				return nil, err
			}

			if execution.QueryExecution.Status.CompletionDateTime != nil {
				results, err := aws.AthenaClient.GetQueryResults(monitoringContext, &athena.GetQueryResultsInput{
					QueryExecutionId: &instance.AthenaQueryId,
				})
				if err != nil {
					return nil, err
				}

				//TODO: Wrap inserting the rows and updating the completedat in a database transaction

				//[1:] here to ignore the first row of results which is the header row
				for _, row := range results.ResultSet.Rows[1:] {
					err = db.InsertUsageReportInstanceProduct(monitoringContext, models.UsageReportInstanceProduct{
						UsageReportInstanceId: instance.Id,
						Product:               *row.Data[0].VarCharValue,
						Value:                 utils.MustParseInt(*row.Data[1].VarCharValue),
					})
					if err != nil {
						return nil, err
					}
				}

				usageReportInstances[i].CompletedAt = utils.TimePtr(time.Now())
				err = db.UpdateUsageReportInstance(monitoringContext, usageReportInstances[i])
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return usageReportInstances, nil
}

func getMissingMonths(usageReports []models.UsageReport, subscription models.Subscription) []missingMonth {
	missingMonths := make([]missingMonth, 0, 2)

	currentMonth := utils.ToMonth(subscription.CreatedAt)

	for currentMonth.Before(utils.ToNextMonth(time.Now())) {
		if !containsMonth(currentMonth, usageReports) {
			missingMonths = append(missingMonths, missingMonth{
				year:  currentMonth.Year(),
				month: int(currentMonth.Month()),
			})
		}

		currentMonth = utils.ToNextMonth(currentMonth)
	}

	return missingMonths
}

func containsMonth(month time.Time, usageReports []models.UsageReport) bool {
	for _, report := range usageReports {
		if month.Year() == report.Year && int(month.Month()) == report.Month {
			return true
		}
	}

	return false
}

func setupTable(monitoringContext *monitoring.Context, report models.UsageReport) error {
	tableName := getTableName(report.SubscriptionId.String(), report.Year, report.Month)
	s3Location := getS3InputLocation(report.SubscriptionId.String(), report.Year, report.Month)

	createTableDDL := fmt.Sprintf(`CREATE EXTERNAL TABLE IF NOT EXISTS %s (
		id STRING,
		occurred_at BIGINT,
		product STRING,
		method STRING,
		path STRING,
		android_id STRING,
		subscription_id STRING
	) ROW FORMAT SERDE 'org.openx.data.jsonserde.JsonSerDe' 
	LOCATION '%s'`, tableName, s3Location)

	ddlResponse, err := aws.AthenaClient.StartQueryExecution(monitoringContext, &athena.StartQueryExecutionInput{
		QueryString: &createTableDDL,
		QueryExecutionContext: &types.QueryExecutionContext{
			Database: &config.GetConfig().AthenaConfig.DatabaseName,
		},
		WorkGroup: &config.GetConfig().AthenaConfig.WorkGroupName,
		ResultConfiguration: &types.ResultConfiguration{
			OutputLocation: utils.StringPtr(getS3OutputLocation()),
		},
	})
	if err != nil {
		monitoringContext.Error("Something went wrong issuing create table statement", zap.Error(err))
		return err
	}

	err = pollForQueryCompletion(monitoringContext, *ddlResponse.QueryExecutionId)
	if err != nil {
		monitoringContext.Error("Something went wrong creating table", zap.Error(err))
		return err
	}

	return nil
}

func createInstanceQuery(monitoringContext *monitoring.Context, report models.UsageReport) (queryId string, err error) {
	tableName := getTableName(report.SubscriptionId.String(), report.Year, report.Month)
	monthlyUsageQuery := fmt.Sprintf("SELECT product, COUNT(1) FROM %s GROUP BY product", tableName)

	queryResponse, err := aws.AthenaClient.StartQueryExecution(monitoringContext, &athena.StartQueryExecutionInput{
		QueryString: &monthlyUsageQuery,
		QueryExecutionContext: &types.QueryExecutionContext{
			Database: &config.GetConfig().AthenaConfig.DatabaseName,
		},
		WorkGroup: &config.GetConfig().AthenaConfig.WorkGroupName,
		ResultConfiguration: &types.ResultConfiguration{
			OutputLocation: utils.StringPtr(getS3OutputLocation()),
		},
	})
	if err != nil {
		return "", err
	}

	return *queryResponse.QueryExecutionId, nil
}

func pollForQueryCompletion(monitoringContext *monitoring.Context, id string) error {
	monitoringContext.Info("Polling for query completion: " + id)
	failedOrCancelled := false
	check := func() error {
		getQueryExecutionOutput, err := aws.AthenaClient.GetQueryExecution(monitoringContext, &athena.GetQueryExecutionInput{
			QueryExecutionId: &id,
		})
		if err != nil {
			return err
		}

		state := getQueryExecutionOutput.QueryExecution.Status.State

		if state == types.QueryExecutionStateFailed || state == types.QueryExecutionStateCancelled {
			failedOrCancelled = true
			return nil
		}

		if state == types.QueryExecutionStateSucceeded {
			return nil
		}

		return errors.New("query is still queued or running")
	}

	err := backoff.Retry(check, &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.2,
		MaxInterval:         1 * time.Second,
		MaxElapsedTime:      5 * time.Second,
		Stop:                -1,
		Clock:               backoff.SystemClock,
	})

	if failedOrCancelled {
		return fmt.Errorf("athena query %s was failed or cancelled", id)
	}

	return err
}

func getTableName(subscriptionId string, year int, month int) string {
	return fmt.Sprintf("usage_report_%s_%s",
		strings.ReplaceAll(subscriptionId, "-", "_"), utils.GetMonth(year, month).Format("2006_01"))
}

func getS3InputLocation(subscriptionId string, year int, month int) string {
	return fmt.Sprintf("s3://%s/%s/%s/", config.GetConfig().AthenaConfig.InputBucketName,
		subscriptionId, utils.GetMonth(year, month).Format("2006/01"))
}

func getS3OutputLocation() string {
	return fmt.Sprintf("s3://%s/", config.GetConfig().AthenaConfig.OutputBucketName)
}
