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

func TestCreateSubscriptionWithoutAPIKeyReturns401(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	resp, err := apiClient.PostSubscriptions(context.Background(), api.PostSubscriptionsJSONRequestBody{
		AccountId: uuid.MustParse("be372162-c0a0-4903-a9e1-a0b372bb1de9"),
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 401, resp.StatusCode)
}

func TestCreateSubscriptionWithInvalidAPIKeyReturns401(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	resp, err := apiClient.PostSubscriptions(context.Background(), api.PostSubscriptionsJSONRequestBody{
		AccountId: uuid.MustParse("be372162-c0a0-4903-a9e1-a0b372bb1de9"),
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", "Bearer 12345")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 401, resp.StatusCode)
}

func TestCreateSubscriptionWithApiKeyWithoutPermissionReturns403(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)
	helper.RunTestSetupScript("api-keys.sql")

	resp, err := apiClient.PostSubscriptions(context.Background(), api.PostSubscriptionsJSONRequestBody{
		AccountId: uuid.MustParse("be372162-c0a0-4903-a9e1-a0b372bb1de9"),
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", "Bearer valid-key-no-permission")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 403, resp.StatusCode)
}

func TestCreateSubscriptionDuplicateReturns409(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)
	helper.RunTestSetupScript("api-keys.sql")
	helper.RunTestSetupScript("existing-subscription.sql")

	resp, err := apiClient.PostSubscriptions(context.Background(), api.PostSubscriptionsJSONRequestBody{
		AccountId: uuid.MustParse("be372162-c0a0-4903-a9e1-a0b372bb1de9"),
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", "Bearer valid-key-with-permission")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 409, resp.StatusCode)
}

func TestCreateSubscriptionReturns201(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)
	helper.RunTestSetupScript("api-keys.sql")

	resp, err := apiClient.PostSubscriptions(context.Background(), api.PostSubscriptionsJSONRequestBody{
		AccountId: uuid.MustParse("be372162-c0a0-4903-a9e1-a0b372bb1de9"),
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", "Bearer valid-key-with-permission")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 201, resp.StatusCode)
	require.True(t, strings.HasPrefix(resp.Header.Get("Location"), "/subscriptions/"))
	_, err = uuid.Parse(strings.Replace(resp.Header.Get("Location"), "/subscriptions/", "", 1))
	require.Nil(t, err)
}
