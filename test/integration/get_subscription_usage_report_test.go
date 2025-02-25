package integration_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"net/http"
	"subscriptions/test/integration/helper"
	"testing"
)

func TestGetSubscriptionUsageWithoutValidSubscriptionReturns404(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	resp, err := apiClient.GetSubscriptionsSubscriptionIdUsageReportsUsageReportId(context.Background(), "c683d6cd-df69-40aa-b268-58e7237e3225", "d456d6cd-df69-40aa-b268-58e1234e3225", func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzdWItYmxhYmxhIiwibmFtZSI6IlNvbWVib2R5IiwiaWF0IjoxMjM0NTYsInN1YnNjcmlwdGlvbl9pZCI6ImM2ODNkNmNkLWRmNjktNDBhYS1iMjY4LTU4ZTcyMzdlMzIyNSIsImFuZHJvaWRfaWQiOiIwN2ZmMDBlNC1jMWE1LTQ2ODMtOWZjYi02MTNhNzM0ZDhkM2YiLCJhY2NvdW50X2lkIjoiYmUzNzIxNjItYzBhMC00OTAzLWE5ZTEtYTBiMzcyYmIxZGU5In0.xVyEIES9mZlwDIbWQYkIrpZ2viNSfI_bgRZ4pQjqBA4")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 404, resp.StatusCode)
}

func TestGetSubscriptionUsageWithValidSubscriptionReturns200(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)
	helper.RunTestSetupScript("existing-usage-report.sql")

	subscriptionId := "c015ce36-76df-4f3f-9352-5daea102d150"

	resp, err := apiClient.GetSubscriptionsSubscriptionIdUsageReportsUsageReportId(context.Background(), subscriptionId, "7a0e62ee-2685-4220-b3b6-3dd0e0c059b9", func(ctx context.Context, req *http.Request) error {
		req.Header.Add("X-Api-Key", "Bearer valid-key-no-permission")
		req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzdWItYmxhYmxhIiwibmFtZSI6IlNvbWVib2R5IiwiaWF0IjoxMjM0NTYsInN1YnNjcmlwdGlvbl9pZCI6ImM2ODNkNmNkLWRmNjktNDBhYS1iMjY4LTU4ZTcyMzdlMzIyNSIsImFuZHJvaWRfaWQiOiIwN2ZmMDBlNC1jMWE1LTQ2ODMtOWZjYi02MTNhNzM0ZDhkM2YiLCJhY2NvdW50X2lkIjoiYmUzNzIxNjItYzBhMC00OTAzLWE5ZTEtYTBiMzcyYmIxZGU5In0.xVyEIES9mZlwDIbWQYkIrpZ2viNSfI_bgRZ4pQjqBA4")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, resp.StatusCode)
}

func TestGetSubscriptionUsageWithoutJwtReturns401(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	resp, err := apiClient.GetSubscriptionsSubscriptionIdUsageReportsUsageReportId(context.Background(), "be372162-c0a0-4903-a9e1-a0b372bb1de9", "d456d6cd-df69-40aa-b268-58e1234e3225", func(ctx context.Context, req *http.Request) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 401, resp.StatusCode)
}
