package services

import (
	uuid2 "github.com/google/uuid"
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
