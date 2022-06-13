package integration_test

import (
	"net/http"
	"subscriptions/src/api"
)

var apiClient = &api.Client{
	Server: "http://localhost:8020",
	Client: http.DefaultClient,
}
