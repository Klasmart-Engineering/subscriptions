package subscription_types_test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"subscriptions/src/models"
	"subscriptions/test/integration/helper"
	"testing"
)

func TestLogAccountActionInvalidBodyReturns400(t *testing.T) {
	helper.ResetDatabase()

	resp, err := http.Post("http://localhost:8020/log-action", "application/json", bytes.NewBuffer([]byte("malformed")))
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 400, resp.StatusCode)
}

func TestInactiveSubscriptionReturnsUnsuccessfulResponse(t *testing.T) {
	helper.ResetDatabase()
	helper.RunTestSql("inactive-subscription.sql")

	var request = models.SubscriptionAccountAction{
		SubscriptionId: "2f797c16-053e-41ab-b40d-24356480e61e",
		ActionType:     "API Call",
		UsageAmount:    1,
		Product:        "Test Product",
		UserId:         "d89124a3-c20d-40fb-82ed-5038dd2aadf2",
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post("http://localhost:8020/log-action", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, resp.StatusCode)

	var responseBody models.LogResponse
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		t.Fatal(err)
	}

	var expectedResponse = models.LogResponse{Success: false, Details: "BLOCKED. Subscription not active", Count: 21, Limit: 30}

	require.Equal(t, expectedResponse, responseBody)
}
