package models

import (
	"errors"
	uuid2 "github.com/google/uuid"
)

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

type Subscription struct {
	Id        uuid2.UUID
	AccountId uuid2.UUID
	State     SubscriptionState
}

type SubscriptionState int

var Active SubscriptionState = 1
var Disabled SubscriptionState = 2
var Deleted SubscriptionState = 3

func SubscriptionStateFromString(value string) (SubscriptionState, error) {
	switch value {
	case "active":
		return Active, nil
	case "disabled":
		return Disabled, nil
	case "deleted":
		return Deleted, nil
	default:
		return Active, errors.New("unknown subscription state: " + value)
	}
}

func (ss SubscriptionState) String() string {
	switch ss {
	case Active:
		return "active"
	case Disabled:
		return "disabled"
	case Deleted:
		return "deleted"
	default:
		return ""
	}
}
