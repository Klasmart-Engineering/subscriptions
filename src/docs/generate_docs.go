package docs

import (
	"subscriptions/src/models"
)

// swagger:route GET /healthcheck application-healthcheck healthcheckEndpoint
// Application Healthcheck. Checking database connection and used by readiness probe
// responses:
//   200: Healthcheck

// Healthcheck Response containing status information.
// swagger:response Healthcheck
type healthcheckResponseWrapper struct {
	// in:body
	Body models.Healthcheck
}

// swagger:route GET /liveness application-liveness livenessEndpoint
// Application Liveness. Checking the application is responsive. Used by the liveness probe
// responses:
//   200: Healthcheck

// Liveness Response containing status information.
// swagger:response Liveness
type livenessResponseWrapper struct {
	// in:body
	Body models.Healthcheck
}

// swagger:route GET /subscription-types subscription-types subscriptionTypesEndpoint
// List the type of subscriptions.
// responses:
//   200: SubscriptionTypeList

// List of subscription types.
// swagger:response subscriptionTypes
type subscriptionTypesResponseWrapper struct {
	// in:body
	Body models.SubscriptionTypeList
}

// swagger:route GET /subscription-actions subscription-actions subscriptionActionsEndpoint
// List the action types for subscriptions.
// responses:
//   200: SubscriptionActionList

// List of subscription actions.
// swagger:response subscriptionTypes
type subscriptionActionsResponseWrapper struct {
	// in:body
	Body models.SubscriptionActionList
}

// swagger:route GET /subscription/{accountID}  getOrCreateSubscriptionEndpoint
// Get or create subscription
// responses:
//   200: models.SubscriptionResponse

// Response for evaluate subscriptions
// swagger:response
type getSubscriptionResponsesWrapper struct {
	// in:body
	Body models.SubscriptionResponse
}

// swagger:parameters getOrCreateSubscriptionEndpoint
type getSubscriptionsWrapper struct {
	// Request containing information about product to add
	// in:body
}

// swagger:route POST /deactivate/{id} deactivate-subscription deactivateSubscriptionEndpoint
// Log actions with API
// responses:
//   200: ProductResponse

// Response containing details about product addition.
// swagger:response ProductResponse
type deactivateSubscriptionResponsesWrapper struct {
	// in:body
	Body models.GenericResponse
}

// swagger:parameters deactivateSubscriptionEndpoint
type deactivateSubscriptionWrapper struct {
	// Request containing information about product to add
	// in:body
	Body string
}
