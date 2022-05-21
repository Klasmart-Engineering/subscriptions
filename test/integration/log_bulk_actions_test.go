package integration_test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"subscriptions/src/models"
	"subscriptions/test/integration/helper"
	"testing"
)

func TestLogBulkActionsInvalidBodyReturns400(t *testing.T) {
	helper.ResetDatabase()

	resp, err := http.Post("http://localhost:8020/log-actions", "application/json", bytes.NewBuffer([]byte("malformed")))
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 400, resp.StatusCode)
}

func TestLogBulkActionsPersistsAllEntriesToDatabase(t *testing.T) {
	helper.ResetDatabase()
	helper.RunTestSetupScript("bulk-actions-log.sql")

	var request = models.SubscriptionAccountActionList{
		Actions: []models.SubscriptionAccountAction{
			{
				SubscriptionId:       "2f797c16-053e-41ab-b40d-24356480e61e",
				ActionType:           "API Call",
				UsageAmount:          1,
				Product:              "Test Product",
				InteractionTimeEpoch: "1653085761",
			},
			{
				SubscriptionId:       "4c7e63ee-43a9-486d-ae38-d3e086593613",
				ActionType:           "API Call",
				UsageAmount:          2,
				Product:              "Test Product",
				InteractionTimeEpoch: "1653085761",
			},
			{
				SubscriptionId:       "5859fc2f-9eed-4e09-b653-5a63d3b100c0",
				ActionType:           "API Call",
				UsageAmount:          3,
				Product:              "Test Product",
				InteractionTimeEpoch: "1653085761",
			},
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post("http://localhost:8020/log-actions", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, resp.StatusCode)

	helper.AssertExactlyOneRowMatchesWithBackoff(t, `SELECT COUNT(1) FROM subscription_account_log WHERE 
													subscription_id = '2f797c16-053e-41ab-b40d-24356480e61e' AND 
													action_type = 'API Call' AND 
													usage = 1 AND 
													product_name = 'Test Product' AND 
													interaction_at = to_timestamp(1653085761)`)
	helper.AssertExactlyOneRowMatchesWithBackoff(t, `SELECT COUNT(1) FROM subscription_account_log WHERE 
													subscription_id = '4c7e63ee-43a9-486d-ae38-d3e086593613' AND 
													action_type = 'API Call' AND 
													usage = 2 AND 
													product_name = 'Test Product' AND 
													interaction_at = to_timestamp(1653085761)`)
	helper.AssertExactlyOneRowMatchesWithBackoff(t, `SELECT COUNT(1) FROM subscription_account_log WHERE 
													subscription_id = '5859fc2f-9eed-4e09-b653-5a63d3b100c0' AND 
													action_type = 'API Call' AND 
													usage = 3 AND 
													product_name = 'Test Product' AND 
													interaction_at = to_timestamp(1653085761)`)
}
