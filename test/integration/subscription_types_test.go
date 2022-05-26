package integration_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"subscriptions/src/models"
	"subscriptions/test/integration/helper"
	"testing"
)

func TestGetSubscriptionTypesReturnsCorrectTypes(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	resp, err := http.Get("http://localhost:8020/subscription-types")
	if err != nil {
		t.Fatal(err)
	}

	var subscriptionTypes models.SubscriptionTypeList
	err = json.NewDecoder(resp.Body).Decode(&subscriptionTypes)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, resp.StatusCode, 200)

	var expectedSubscriptionTypes = models.SubscriptionTypeList{
		Subscriptions: []models.SubscriptionType{
			{
				ID:   2,
				Name: "Uncapped",
			},
			{
				ID:   1,
				Name: "Capped",
			},
		},
	}

	require.Equal(t, expectedSubscriptionTypes, subscriptionTypes)
}
