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

func TestGetSubscriptionUsagesWithoutValidSubscriptionReturns404(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	resp, err := apiClient.GetSubscriptionsSubscriptionIdUsageReports(context.Background(), "c683d6cd-df69-40aa-b268-58e7237e3225", func(ctx context.Context, req *http.Request) error {
		req.Header.Add("X-Api-Key", "Bearer valid-key-no-permission")
		req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzdWItYmxhYmxhIiwibmFtZSI6IlNvbWVib2R5IiwiaWF0IjoxMjM0NTYsInN1YnNjcmlwdGlvbl9pZCI6ImM2ODNkNmNkLWRmNjktNDBhYS1iMjY4LTU4ZTcyMzdlMzIyNSIsImFuZHJvaWRfaWQiOiIwN2ZmMDBlNC1jMWE1LTQ2ODMtOWZjYi02MTNhNzM0ZDhkM2YiLCJhY2NvdW50X2lkIjoiYmUzNzIxNjItYzBhMC00OTAzLWE5ZTEtYTBiMzcyYmIxZGU5In0.xVyEIES9mZlwDIbWQYkIrpZ2viNSfI_bgRZ4pQjqBA4")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 404, resp.StatusCode)
}

func TestGetSubscriptionUsagesWithValidSubscriptionReturns200(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)
	helper.RunTestSetupScript("api-keys.sql")

	response, err := apiClient.PostSubscriptions(context.Background(), api.PostSubscriptionsJSONRequestBody{
		AccountId: uuid.MustParse("be372162-c0a0-4903-a9e1-a0b372bb1de9"),
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("X-Api-Key", "Bearer valid-key-with-permission")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	subscriptionId := strings.Replace(response.Header.Get("Location"), "/subscriptions/", "", 1)
	resp, err := apiClient.GetSubscriptionsSubscriptionIdUsageReports(context.Background(), subscriptionId, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("X-Api-Key", "Bearer valid-key-no-permission")
		req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzdWItYmxhYmxhIiwibmFtZSI6IlNvbWVib2R5IiwiaWF0IjoxMjM0NTYsInN1YnNjcmlwdGlvbl9pZCI6ImM2ODNkNmNkLWRmNjktNDBhYS1iMjY4LTU4ZTcyMzdlMzIyNSIsImFuZHJvaWRfaWQiOiIwN2ZmMDBlNC1jMWE1LTQ2ODMtOWZjYi02MTNhNzM0ZDhkM2YiLCJhY2NvdW50X2lkIjoiYmUzNzIxNjItYzBhMC00OTAzLWE5ZTEtYTBiMzcyYmIxZGU5In0.xVyEIES9mZlwDIbWQYkIrpZ2viNSfI_bgRZ4pQjqBA4")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 201, response.StatusCode)
	require.Equal(t, 200, resp.StatusCode)
}

func TestGetSubscriptionUsagesWithoutJwtReturns401(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	resp, err := apiClient.GetSubscriptionsSubscriptionIdUsageReports(context.Background(), "be372162-c0a0-4903-a9e1-a0b372bb1de9")
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 401, resp.StatusCode)
}
