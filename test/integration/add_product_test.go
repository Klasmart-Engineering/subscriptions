package integration_test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"subscriptions/src/models"
	"subscriptions/test/integration/helper"
	"testing"
)

func TestAddProductInvalidBodyReturns400(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	resp, err := http.Post("http://localhost:8020/add-product", "application/json", bytes.NewBuffer([]byte("malformed")))
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 400, resp.StatusCode)
}

func TestAddProductToSubscriptionSucceeds(t *testing.T) {
	helper.ResetDatabase()
	helper.RunTestSetupScript("add-product-to-subscription.sql")
	helper.WaitForHealthcheck(t)

	var request = models.AddProduct{
		SubscriptionId: "2f797c16-053e-41ab-b40d-24356480e61e",
		Product:        "My Product",
		Type:           "Capped",
		Threshold:      50,
		Action:         "UserLogin",
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post("http://localhost:8020/add-product", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 200, resp.StatusCode)

	var responseBody models.ProductResponse
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		t.Fatal(err)
	}

	var expectedResponse = models.ProductResponse{Details: "Successfully added product"}

	require.Equal(t, expectedResponse, responseBody)

	helper.AssertExactlyOneRowMatchesWithBackoff(t, `SELECT COUNT(1) FROM subscription_account_product WHERE 
													subscription_id = '2f797c16-053e-41ab-b40d-24356480e61e' AND 
													product = 'My Product' AND 
													type = 'Capped' AND 
													threshold = 50 AND 
													action = 'UserLogin'`)
}
