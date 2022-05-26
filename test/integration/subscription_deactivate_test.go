package integration_test

import (
	"bytes"
	"encoding/json"
	uuid2 "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"subscriptions/src/models"
	"subscriptions/test/integration/helper"
	"testing"
)

func TestDeactivateSubscriptionReturnsSubscriptionUuid(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	accountId, err := uuid2.NewUUID()
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

	id := responseBody.SubscriptionId
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.Post("http://localhost:8020/deactivate/"+id, "application/json", bytes.NewBuffer([]byte("")))

	if err != nil {
		t.Fatal(err)
	}

	var deactivateResponseBody models.GenericResponse

	err = json.NewDecoder(resp.Body).Decode(&deactivateResponseBody)
	if err != nil {
		t.Fatal(err)
	}

	var expectedResponse = models.GenericResponse{Details: "Subscription deactivated."}

	require.Equal(t, expectedResponse, deactivateResponseBody)
}
