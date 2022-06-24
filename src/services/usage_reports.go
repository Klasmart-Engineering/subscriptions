package services

import (
	"github.com/aws/aws-sdk-go-v2/service/athena"
	uuid2 "github.com/google/uuid"
	"subscriptions/src/aws"
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
	//TODO: Create table if needed
	//		Create query
	//		Write into usage report instance table
	//		Bucket names and all that jazz need parameterising from config rather than hardcoding!
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

	for _, instance := range usageReportInstances {
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
				for _, row := range results.ResultSet.Rows {
					err = db.InsertUsageReportInstanceProduct(monitoringContext, models.UsageReportInstanceProduct{
						UsageReportInstanceId: instance.Id,
						Product:               *row.Data[0].VarCharValue,
						Value:                 utils.MustParseInt(*row.Data[1].VarCharValue),
					})
					if err != nil {
						return nil, err
					}
				}

				instance.CompletedAt = utils.TimePtr(time.Now())
				err = db.UpdateUsageReportInstance(monitoringContext, instance)
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
