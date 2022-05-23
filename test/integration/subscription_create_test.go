package subscription_types_test

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

func TestCreateSubscriptionReturnsSubscriptionUuid(t *testing.T) {
	helper.ResetDatabase()
	resp, err := http.Post("http://localhost:8020/create-subscription", "application/json", bytes.NewBuffer([]byte("")))
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

}
