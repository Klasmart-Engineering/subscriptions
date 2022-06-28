package integration_test

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"subscriptions/src/api"
	"subscriptions/test/integration/helper"
	"testing"
)

func TestGetSubscriptionByAccountIdWithoutAPIKeyReturns401(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	resp, err := apiClient.GetSubscriptions(context.Background(), &api.GetSubscriptionsParams{
		AccountId: "be372162-c0a0-4903-a9e1-a0b372bb1de9",
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 401, resp.StatusCode)
}

func TestGetSubscriptionByAccountIdWithInvalidAPIKeyReturns401(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	resp, err := apiClient.GetSubscriptions(context.Background(), &api.GetSubscriptionsParams{
		AccountId: "be372162-c0a0-4903-a9e1-a0b372bb1de9",
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("X-Api-Key", "Bearer 12345")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 401, resp.StatusCode)
}

func TestGetSubscriptionByAccountIdWithApiKeyWithoutPermissionReturns403(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)
	helper.RunTestSetupScript("api-keys.sql")

	resp, err := apiClient.GetSubscriptions(context.Background(), &api.GetSubscriptionsParams{
		AccountId: "be372162-c0a0-4903-a9e1-a0b372bb1de9",
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("X-Api-Key", "Bearer valid-key-no-permission")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 403, resp.StatusCode)
}

func TestGetSubscriptionByAccountIdWithApiKeyNonExistentSubscriptionReturnsEmpty200(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)
	helper.RunTestSetupScript("api-keys.sql")

	resp, err := apiClient.GetSubscriptions(context.Background(), &api.GetSubscriptionsParams{
		AccountId: "be372162-c0a0-4903-a9e1-a0b372bb1de9",
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("X-Api-Key", "Bearer valid-key-with-permission")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, resp.StatusCode)

	var responseBody []api.Subscription
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		t.Fatal(err)
	}

	require.Empty(t, responseBody)
}

func TestGetSubscriptionByAccountIdWithApiKeyReturns200(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)
	helper.RunTestSetupScript("api-keys.sql")
	helper.RunTestSetupScript("existing-subscription.sql")

	resp, err := apiClient.GetSubscriptions(context.Background(), &api.GetSubscriptionsParams{
		AccountId: "be372162-c0a0-4903-a9e1-a0b372bb1de9",
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("X-Api-Key", "Bearer valid-key-with-permission")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, resp.StatusCode)

	var responseBody []api.Subscription
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, responseBody, 1)
	require.Equal(t, api.Subscription{
		AccountId: uuid.MustParse("be372162-c0a0-4903-a9e1-a0b372bb1de9"),
		Id:        uuid.MustParse("c683d6cd-df69-40aa-b268-58e7237e3225"),
		State:     "disabled",
		CreatedOn: 1656374400,
	}, responseBody[0])
}

func TestGetSubscriptionByAccountIdWithJwtReturns200(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)
	helper.RunTestSetupScript("existing-subscription.sql")

	resp, err := apiClient.GetSubscriptions(context.Background(), &api.GetSubscriptionsParams{
		AccountId: "be372162-c0a0-4903-a9e1-a0b372bb1de9",
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", "Bearer eyJhbGciOiAiSFMyNTYiLCJ0eXAiOiAiSldUIn0.eyJzdWIiOiAic3ViLWJsYWJsYSIsIm5hbWUiOiAiU29tZWJvZHkiLCJpYXQiOiAxMjM0NTYsInN1YnNjcmlwdGlvbl9pZCI6ICJjNjgzZDZjZC1kZjY5LTQwYWEtYjI2OC01OGU3MjM3ZTMyMjUiLCJhbmRyb2lkX2lkIjogIjA3ZmYwMGU0LWMxYTUtNDY4My05ZmNiLTYxM2E3MzRkOGQzZiJ9.aW52YWxpZCBzaWduYXR1cmU")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, resp.StatusCode)

	var responseBody []api.Subscription
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, responseBody, 1)
	require.Equal(t, api.Subscription{
		AccountId: uuid.MustParse("be372162-c0a0-4903-a9e1-a0b372bb1de9"),
		Id:        uuid.MustParse("c683d6cd-df69-40aa-b268-58e7237e3225"),
		State:     "disabled",
		CreatedOn: 1656374400,
	}, responseBody[0])
}

func TestGetSubscriptionByAccountIdNonMatchingAccountIdWithJwtReturnsEmpty200(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)
	helper.RunTestSetupScript("existing-subscription.sql")

	resp, err := apiClient.GetSubscriptions(context.Background(), &api.GetSubscriptionsParams{
		AccountId: "be372162-c0a0-4903-a9e1-a0b372bb1de9",
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", "Bearer eyJhbGciOiAiSFMyNTYiLCJ0eXAiOiAiSldUIn0.eyJzdWIiOiAic3ViLWJsYWJsYSIsIm5hbWUiOiAiU29tZWJvZHkiLCJpYXQiOiAxMjM0NTYsInN1YnNjcmlwdGlvbl9pZCI6ICJhOWRlOTNmYy0yZDEzLTQ0ZGQtOTI3Mi1kYTdmOGMxN2QxNTUiLCJhbmRyb2lkX2lkIjogIjA3ZmYwMGU0LWMxYTUtNDY4My05ZmNiLTYxM2E3MzRkOGQzZiIsICJhY2NvdW50X2lkIjogIjA3ZmYwMGU0LWMxYTUtNDY4My05ZmNiLTYxM2E3MzRkOGQzZiJ9.aW52YWxpZCBzaWduYXR1cmU")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, resp.StatusCode)

	var responseBody []api.Subscription
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		t.Fatal(err)
	}

	require.Empty(t, responseBody)
}
