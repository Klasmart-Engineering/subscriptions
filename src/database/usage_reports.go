package db

import (
	uuid2 "github.com/google/uuid"
	"subscriptions/src/models"
	"subscriptions/src/monitoring"
)

func GetUsageReportsForSubscription(monitoringContext *monitoring.Context, subscriptionId uuid2.UUID) ([]models.UsageReport, error) {
	var result []models.UsageReport

	err := dbConnection.SelectContext(monitoringContext, &result, `
		SELECT * FROM usage_report WHERE subscription_id = $1`, subscriptionId)

	return result, err
}

func InsertUsageReport(monitoringContext *monitoring.Context, usageReport models.UsageReport) error {
	_, err := dbConnection.ExecContext(monitoringContext, `
		INSERT INTO usage_report (id, subscription_id, year, month) VALUES ($1, $2, $3, $4)`,
		usageReport.Id, usageReport.SubscriptionId, usageReport.Year, usageReport.Month)

	return err
}

func GetUsageReportInstances(monitoringContext *monitoring.Context, usageReportId uuid2.UUID) ([]models.UsageReportInstance, error) {
	var result []models.UsageReportInstance

	err := dbConnection.SelectContext(monitoringContext, &result, `
		SELECT * FROM usage_report_instance WHERE usage_report_id = $1`, usageReportId)

	return result, err
}

func InsertUsageReportInstance(monitoringContext *monitoring.Context, usageReportInstance models.UsageReportInstance) error {
	_, err := dbConnection.ExecContext(monitoringContext, `
		INSERT INTO usage_report_instance (id, usage_report_id, requested_at, athena_query_id, completed_at) 
		VALUES ($1, $2, $3, $4, $5)`,
		usageReportInstance.Id, usageReportInstance.UsageReportId, usageReportInstance.RequestedAt, usageReportInstance.AthenaQueryId, usageReportInstance.CompletedAt)

	return err
}

func UpdateUsageReportInstance(monitoringContext *monitoring.Context, usageReportInstance models.UsageReportInstance) error {
	_, err := dbConnection.ExecContext(monitoringContext, `
		UPDATE usage_report_instance SET usage_report_id = $1, requested_at = $2, athena_query_id = $3, completed_at = $4  
		WHERE id = $5`,
		usageReportInstance.UsageReportId, usageReportInstance.RequestedAt, usageReportInstance.AthenaQueryId, usageReportInstance.CompletedAt, usageReportInstance.Id)

	return err
}

func GetUsageReportInstanceProducts(monitoringContext *monitoring.Context, usageReportInstance uuid2.UUID) ([]models.UsageReportInstanceProduct, error) {
	var result []models.UsageReportInstanceProduct

	err := dbConnection.SelectContext(monitoringContext, &result, `
		SELECT * FROM usage_report_instance_product WHERE usage_report_instance_id = $1`, usageReportInstance)

	return result, err
}

func InsertUsageReportInstanceProduct(monitoringContext *monitoring.Context, usageReportInstanceProduct models.UsageReportInstanceProduct) error {
	_, err := dbConnection.ExecContext(monitoringContext, `
		INSERT INTO usage_report_instance_product (usage_report_instance_id, product, value) 
		VALUES ($1, $2, $3)`,
		usageReportInstanceProduct.UsageReportInstanceId, usageReportInstanceProduct.Product, usageReportInstanceProduct.Value)

	return err
}
