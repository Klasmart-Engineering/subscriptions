package db

import (
	"subscriptions/src/models"
	"subscriptions/src/monitoring"
)

func GetApiKeys(monitoringContext *monitoring.Context) ([]models.ApiKeyPermission, error) {
	var result []models.ApiKeyPermission

	err := dbConnection.SelectContext(monitoringContext, &result, `
		SELECT ak.owner, ak.api_key, akp.permission FROM api_key ak 
			LEFT JOIN api_key_permission akp ON akp.owner = ak.owner`)

	return result, err
}
