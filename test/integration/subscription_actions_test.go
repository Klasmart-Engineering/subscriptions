package integration_test

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"subscriptions/src/api"
	"subscriptions/test/integration/helper"
	"testing"
)

func TestGetSubscriptionActionsReturnsCorrectActions(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	resp, err := apiClient.GetSubscriptionActions(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	var subscriptionActions []api.SubscriptionAction
	err = json.NewDecoder(resp.Body).Decode(&subscriptionActions)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, resp.StatusCode, 200)

	var expectedSubscriptionActions = []api.SubscriptionAction{
		{
			Name:        "API Call",
			Description: "User interaction with public API Gateway",
			Unit:        "HTTP Requests",
		},
	}

	require.Equal(t, expectedSubscriptionActions, subscriptionActions)
}
