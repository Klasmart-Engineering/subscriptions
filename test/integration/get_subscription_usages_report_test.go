package integration_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"subscriptions/test/integration/helper"
	"testing"
)

func TestGetSubscriptionUsagesWithoutValidSubscriptionReturns404(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	resp, err := apiClient.GetSubscriptionsSubscriptionIdUsageReports(context.Background(), "c683d6cd-df69-40aa-b268-58e7237e3225")
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 404, resp.StatusCode)
}

func TestGetSubscriptionUsagesWithValidSubscriptionReturns200(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)
	helper.RunTestSetupScript("existing-subscription.sql")

	resp, err := apiClient.GetSubscriptionsSubscriptionIdUsageReports(context.Background(), "c683d6cd-df69-40aa-b268-58e7237e3225")
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, resp.StatusCode)
}
