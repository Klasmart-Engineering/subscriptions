package models

import (
	"fmt"
	"net/http"
)

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

func (*ProductResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (*SubscriptionResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (*GenericResponse) Render(w http.ResponseWriter, r *http.Request) error {
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

type Healthcheck struct {
	Up      bool   `json:"up"`
	Details string `json:"details"`
}

func (*Healthcheck) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type ProductResponse struct {
	Details string `json:"details"`
}

type GenericResponse struct {
	Details string `json:"details"`
}

type SubscriptionResponse struct {
	SubscriptionId string `json:"subscriptionId"`
	Active         bool   `json:"active"`
}

type SubscriptionChange struct {
	SubscriptionId int  `json:"subscriptionId"`
	Active         bool `json:"active"`
}
