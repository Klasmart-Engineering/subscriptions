package integration_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"subscriptions/src/api"
	"subscriptions/test/integration/helper"
	"testing"
)

func TestPatchSubscriptionReturns200(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)
	helper.RunTestSetupScript("api-keys.sql")

	resp, err := apiClient.PostSubscriptions(context.Background(), api.PostSubscriptionsJSONRequestBody{
		AccountId: uuid.MustParse("be372162-c0a0-4903-a9e1-a0b372bb1de9"),
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("X-Api-Key", "Bearer valid-key-with-permission")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	subscriptionId := strings.Replace(resp.Header.Get("Location"), "/subscriptions/", "", 1)
	responsePatch, err := apiClient.PatchSubscriptionsSubscriptionId(context.Background(), subscriptionId, api.PatchSubscriptionsSubscriptionIdJSONRequestBody{
		State: "active",
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzdWItYmxhYmxhIiwibmFtZSI6IlNvbWVib2R5IiwiaWF0IjoxMjM0NTYsInN1YnNjcmlwdGlvbl9pZCI6ImM2ODNkNmNkLWRmNjktNDBhYS1iMjY4LTU4ZTcyMzdlMzIyNSIsImFuZHJvaWRfaWQiOiIwN2ZmMDBlNC1jMWE1LTQ2ODMtOWZjYi02MTNhNzM0ZDhkM2YiLCJhY2NvdW50X2lkIjoiYmUzNzIxNjItYzBhMC00OTAzLWE5ZTEtYTBiMzcyYmIxZGU5In0.xVyEIES9mZlwDIbWQYkIrpZ2viNSfI_bgRZ4pQjqBA4")
		req.Header.Add("X-Api-Key", "Bearer valid-key-with-permission")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 201, resp.StatusCode)
	require.Equal(t, 200, responsePatch.StatusCode)
	require.True(t, strings.HasPrefix(resp.Header.Get("Location"), "/subscriptions/"))
	_, err = uuid.Parse(subscriptionId)
	require.Nil(t, err)
}
