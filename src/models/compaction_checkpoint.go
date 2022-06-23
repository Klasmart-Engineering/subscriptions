package models

import (
	uuid2 "github.com/google/uuid"
	"time"
)

type CompactionCheckpoint struct {
	SubscriptionId uuid2.UUID
	SucceededAt    *time.Time
	FailedAt       *time.Time
}
