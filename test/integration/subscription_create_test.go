package integration_test

import (
	"encoding/json"
	uuid2 "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"subscriptions/src/models"
	"subscriptions/test/integration/helper"
	"testing"
)

func TestCreateSubscriptionWithNewAccountIdReturnsActiveSubscription(t *testing.T) {
	helper.WaitForHealthcheck(t)

	accountId, err := uuid2.NewUUID()
	helper.ResetDatabase()
	resp, err := http.Get("http://localhost:8020/subscription/" + accountId.String())
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, resp.StatusCode)

	var responseBody models.SubscriptionResponse
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		t.Fatal(err)
	}

	_, err = uuid2.Parse(responseBody.SubscriptionId)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, true, responseBody.Active)
}
