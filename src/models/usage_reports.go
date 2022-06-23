package models

import (
	uuid2 "github.com/google/uuid"
	"time"
)

type UsageReport struct {
	Id             uuid2.UUID
	SubscriptionId uuid2.UUID
	Year           int
	Month          int
}

type UsageReportInstance struct {
	Id            uuid2.UUID
	UsageReportId uuid2.UUID
	RequestedAt   time.Time
	AthenaQueryId string
	CompletedAt   *time.Time
}

type UsageReportInstanceProduct struct {
	UsageReportInstanceId uuid2.UUID
	Product               string
	Value                 int
}
