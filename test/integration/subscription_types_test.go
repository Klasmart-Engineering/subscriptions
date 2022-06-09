package integration_test

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"subscriptions/src/api"
	"subscriptions/test/integration/helper"
	"testing"
)

func TestGetSubscriptionTypesReturnsCorrectTypes(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	resp, err := apiClient.GetSubscriptionTypes(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	var subscriptionTypes []api.SubscriptionType
	err = json.NewDecoder(resp.Body).Decode(&subscriptionTypes)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, resp.StatusCode, 200)

	var expectedSubscriptionTypes = []api.SubscriptionType{
		{
			Id:   2,
			Name: "Uncapped",
		},
		{
			Id:   1,
			Name: "Capped",
		},
	}

	require.Equal(t, expectedSubscriptionTypes, subscriptionTypes)
}
