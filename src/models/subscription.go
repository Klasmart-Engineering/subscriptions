package models

import (
	"fmt"
	"net/http"
)

type EvaluatedSubscription struct {
	SubscriptionId string                         `json:"subscriptionId"`
	Products       []EvaluatedSubscriptionProduct `json:"products"`
	DateFromEpoch  string                         `json:"dateFromEpoch"`
	DateToEpoch    string                         `json:"dateToEpoch"`
}

type EvaluatedSubscriptionProduct struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	UsageAmount int    `json:"usageAmount"`
}

type SubscriptionEvaluationProduct struct {
	Threshold int    `json:"threshold"`
	Name      string `json:"name"`
	Type      string `json:"type"`
}

type SubscriptionEvaluation struct {
	ID                string
	Products          []SubscriptionEvaluationProduct
	LastProcessedTime string
}

type SubscriptionEvaluations struct {
	SubscriptionEvaluations []SubscriptionEvaluation
}

type SubscriptionType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type SubscriptionTypeList struct {
	Subscriptions []SubscriptionType `json:"subscriptions"`
}

func (i *SubscriptionType) Bind(r *http.Request) error {
	if i.Name == "" {
		return fmt.Errorf("name is a required field")
	}
	return nil
}
func (*SubscriptionTypeList) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func (*SubscriptionType) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type SubscriptionAction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Unit        string `json:"unit"`
}
type SubscriptionActionList struct {
	Actions []SubscriptionAction `json:"actions"`
}

func (i *SubscriptionAction) Bind(r *http.Request) error {
	return nil
}
func (*SubscriptionActionList) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func (*SubscriptionAction) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type LogResponse struct {
	Success bool   `json:"success"`
	Details string `json:"details"`
	Count   int    `json:"count"`
	Limit   int    `json:"limit"`
}

type Healthcheck struct {
	Up      bool   `json:"up"`
	Details string `json:"details"`
}

func (*Healthcheck) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (*LogResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type SubscriptionAccountAction struct {
	SubscriptionId string `json:"SubscriptionId"`
	ActionType     string `json:"actionType"`
	UsageAmount    int    `json:"usageAmount"`
	Product        string `json:"product"`
}

type SubscriptionChange struct {
	SubscriptionId int  `json:"subscriptionId"`
	Active         bool `json:"active"`
}
