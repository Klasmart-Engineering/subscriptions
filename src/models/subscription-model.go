package models

type SubscriptionType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type SubscriptionTypeList struct {
	Subscriptions []SubscriptionType `json:"subscriptions"`
}

type SubscriptionAction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Unit        string `json:"unit"`
}
type SubscriptionActionList struct {
	Actions []SubscriptionAction `json:"actions"`
}

type Healthcheck struct {
	Up      bool   `json:"up"`
	Details string `json:"details"`
}
