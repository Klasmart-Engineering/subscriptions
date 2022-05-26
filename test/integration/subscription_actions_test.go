package integration_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"subscriptions/src/models"
	"subscriptions/test/integration/helper"
	"testing"
)

func TestGetSubscriptionActionsReturnsCorrectActions(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	resp, err := http.Get("http://localhost:8020/subscription-actions")
	if err != nil {
		t.Fatal(err)
	}

	var subscriptionActions models.SubscriptionActionList
	err = json.NewDecoder(resp.Body).Decode(&subscriptionActions)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, resp.StatusCode, 200)

	var expectedSubscriptionActions = models.SubscriptionActionList{
		Actions: []models.SubscriptionAction{
			{
				Name:        "API Call",
				Description: "User interaction with public API Gateway",
				Unit:        "HTTP Requests",
			},
			{
				Name:        "UserLogin",
				Description: "User Login Action",
				Unit:        "Account Active",
			},
		},
	}

	require.Equal(t, expectedSubscriptionActions, subscriptionActions)
}
