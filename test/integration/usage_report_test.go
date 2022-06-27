package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"subscriptions/src/api"
	"subscriptions/test/integration/helper"
	"testing"
	"time"
)

func TestUsageReportCanBeRequestedThenPolledForData(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)
	helper.RunTestSetupScript("usage-report.sql")

	resp, err := apiClient.GetSubscriptionsSubscriptionIdUsageReports(context.Background(),
		"14fb4f6e-1298-4ca5-989d-00b56a2c6564",
		func(ctx context.Context, req *http.Request) error {
			req.Header.Add("Authorization", "Bearer eyJhbGciOiAiSFMyNTYiLCJ0eXAiOiAiSldUIn0.eyJzdWIiOiAic3ViLWJsYWJsYSIsIm5hbWUiOiAiU29tZWJvZHkiLCJpYXQiOiAxMjM0NTYsInN1YnNjcmlwdGlvbl9pZCI6ICIxOTY2ZjM2OC01NTI4LTQ1OTEtOTlkMS0zYzQ3NWEwMmIxZjUiLCJhbmRyb2lkX2lkIjogIjA3ZmYwMGU0LWMxYTUtNDY4My05ZmNiLTYxM2E3MzRkOGQzZiIsICJhY2NvdW50X2lkIjogImJlMzcyMTYyLWMwYTAtNDkwMy1hOWUxLWEwYjM3MmJiMWRlOSJ9.aW52YWxpZCBzaWduYXR1cmU")
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, resp.StatusCode)

	var responseBody []api.UsageReports
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		t.Fatal(err)
	}

	require.Greater(t, len(responseBody), 0)

	usageResp, err := apiClient.GetSubscriptionsSubscriptionIdUsageReportsUsageReportId(context.Background(),
		"14fb4f6e-1298-4ca5-989d-00b56a2c6564",
		responseBody[0].Id.String(),
		func(ctx context.Context, req *http.Request) error {
			req.Header.Add("Authorization", "Bearer eyJhbGciOiAiSFMyNTYiLCJ0eXAiOiAiSldUIn0.eyJzdWIiOiAic3ViLWJsYWJsYSIsIm5hbWUiOiAiU29tZWJvZHkiLCJpYXQiOiAxMjM0NTYsInN1YnNjcmlwdGlvbl9pZCI6ICIxOTY2ZjM2OC01NTI4LTQ1OTEtOTlkMS0zYzQ3NWEwMmIxZjUiLCJhbmRyb2lkX2lkIjogIjA3ZmYwMGU0LWMxYTUtNDY4My05ZmNiLTYxM2E3MzRkOGQzZiIsICJhY2NvdW50X2lkIjogImJlMzcyMTYyLWMwYTAtNDkwMy1hOWUxLWEwYjM3MmJiMWRlOSJ9.aW52YWxpZCBzaWduYXR1cmU")
			return nil
		})

	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, usageResp.StatusCode)

	var usageReportResponseBody api.UsageReport
	err = json.NewDecoder(usageResp.Body).Decode(&usageReportResponseBody)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, responseBody[0].Id, usageReportResponseBody.Id)
	require.Equal(t, "not_requested", usageReportResponseBody.State)
	require.Equal(t, responseBody[0].From, usageReportResponseBody.From)
	require.Equal(t, responseBody[0].To, usageReportResponseBody.To)
	require.Equal(t, (*int64)(nil), usageReportResponseBody.ReportCompletedAt)
	require.Equal(t, (*api.UsageReport_Products)(nil), usageReportResponseBody.Products)

	usagePatchResp, err := apiClient.PatchSubscriptionsSubscriptionIdUsageReportsUsageReportId(context.Background(),
		"14fb4f6e-1298-4ca5-989d-00b56a2c6564",
		responseBody[0].Id.String(),
		func(ctx context.Context, req *http.Request) error {
			req.Header.Add("Authorization", "Bearer eyJhbGciOiAiSFMyNTYiLCJ0eXAiOiAiSldUIn0.eyJzdWIiOiAic3ViLWJsYWJsYSIsIm5hbWUiOiAiU29tZWJvZHkiLCJpYXQiOiAxMjM0NTYsInN1YnNjcmlwdGlvbl9pZCI6ICIxOTY2ZjM2OC01NTI4LTQ1OTEtOTlkMS0zYzQ3NWEwMmIxZjUiLCJhbmRyb2lkX2lkIjogIjA3ZmYwMGU0LWMxYTUtNDY4My05ZmNiLTYxM2E3MzRkOGQzZiIsICJhY2NvdW50X2lkIjogImJlMzcyMTYyLWMwYTAtNDkwMy1hOWUxLWEwYjM3MmJiMWRlOSJ9.aW52YWxpZCBzaWduYXR1cmU")
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, usagePatchResp.StatusCode)

	usageResp2, err := apiClient.GetSubscriptionsSubscriptionIdUsageReportsUsageReportId(context.Background(),
		"14fb4f6e-1298-4ca5-989d-00b56a2c6564",
		responseBody[0].Id.String(),
		func(ctx context.Context, req *http.Request) error {
			req.Header.Add("Authorization", "Bearer eyJhbGciOiAiSFMyNTYiLCJ0eXAiOiAiSldUIn0.eyJzdWIiOiAic3ViLWJsYWJsYSIsIm5hbWUiOiAiU29tZWJvZHkiLCJpYXQiOiAxMjM0NTYsInN1YnNjcmlwdGlvbl9pZCI6ICIxOTY2ZjM2OC01NTI4LTQ1OTEtOTlkMS0zYzQ3NWEwMmIxZjUiLCJhbmRyb2lkX2lkIjogIjA3ZmYwMGU0LWMxYTUtNDY4My05ZmNiLTYxM2E3MzRkOGQzZiIsICJhY2NvdW50X2lkIjogImJlMzcyMTYyLWMwYTAtNDkwMy1hOWUxLWEwYjM3MmJiMWRlOSJ9.aW52YWxpZCBzaWduYXR1cmU")
			return nil
		})

	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, usageResp2.StatusCode)

	var usageReportResponseBody2 api.UsageReport
	err = json.NewDecoder(usageResp2.Body).Decode(&usageReportResponseBody2)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, responseBody[0].Id, usageReportResponseBody2.Id)
	require.Equal(t, "processing", usageReportResponseBody2.State)
	require.Equal(t, responseBody[0].From, usageReportResponseBody2.From)
	require.Equal(t, responseBody[0].To, usageReportResponseBody2.To)
	require.Equal(t, (*int64)(nil), usageReportResponseBody2.ReportCompletedAt)
	require.Equal(t, (*api.UsageReport_Products)(nil), usageReportResponseBody2.Products)

	time.Sleep(time.Second * 6)

	usageResp3, err := apiClient.GetSubscriptionsSubscriptionIdUsageReportsUsageReportId(context.Background(),
		"14fb4f6e-1298-4ca5-989d-00b56a2c6564",
		responseBody[0].Id.String(),
		func(ctx context.Context, req *http.Request) error {
			req.Header.Add("Authorization", "Bearer eyJhbGciOiAiSFMyNTYiLCJ0eXAiOiAiSldUIn0.eyJzdWIiOiAic3ViLWJsYWJsYSIsIm5hbWUiOiAiU29tZWJvZHkiLCJpYXQiOiAxMjM0NTYsInN1YnNjcmlwdGlvbl9pZCI6ICIxOTY2ZjM2OC01NTI4LTQ1OTEtOTlkMS0zYzQ3NWEwMmIxZjUiLCJhbmRyb2lkX2lkIjogIjA3ZmYwMGU0LWMxYTUtNDY4My05ZmNiLTYxM2E3MzRkOGQzZiIsICJhY2NvdW50X2lkIjogImJlMzcyMTYyLWMwYTAtNDkwMy1hOWUxLWEwYjM3MmJiMWRlOSJ9.aW52YWxpZCBzaWduYXR1cmU")
			return nil
		})

	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, usageResp3.StatusCode)

	bodyString, _ := io.ReadAll(usageResp3.Body)

	var usageReportResponseBody3 api.UsageReport
	err = json.NewDecoder(bytes.NewBuffer(bodyString)).Decode(&usageReportResponseBody3)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, responseBody[0].Id, usageReportResponseBody3.Id)
	require.Equal(t, "ready", usageReportResponseBody3.State)
	require.Equal(t, responseBody[0].From, usageReportResponseBody3.From)
	require.Equal(t, responseBody[0].To, usageReportResponseBody3.To)
	require.Greater(t, *usageReportResponseBody3.ReportCompletedAt, time.Now().Add(0-(time.Second*10)).Unix())
	require.Equal(t, 54, usageReportResponseBody3.Products.AdditionalProperties["Product A"])
	require.Equal(t, 122, usageReportResponseBody3.Products.AdditionalProperties["Product B"])
}
