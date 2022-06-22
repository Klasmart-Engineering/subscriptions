package db

import (
	"database/sql"
	uuid2 "github.com/google/uuid"
	"subscriptions/src/models"
	"subscriptions/src/monitoring"
)

func GetCompactionCheckpoint(monitoringContext *monitoring.Context, subscriptionId uuid2.UUID) (exists bool, entity models.CompactionCheckpoint, err error) {
	var result models.CompactionCheckpoint

	getErr := dbConnection.GetContext(monitoringContext, &result,
		"SELECT * FROM compaction_checkpoint WHERE subscription_id = $1", subscriptionId)
	if getErr != nil {
		if getErr == sql.ErrNoRows {
			return false, result, nil
		}

		return false, result, getErr
	}

	return true, result, nil
}

func UpsertCompactionCheckpoint(monitoringContext *monitoring.Context, checkpoint models.CompactionCheckpoint) error {
	if exists, _, _ := GetCompactionCheckpoint(monitoringContext, checkpoint.SubscriptionId); exists {
		_, err := dbConnection.ExecContext(monitoringContext,
			"UPDATE compaction_checkpoint SET succeeded_at = $1, failed_at = $2 WHERE subscription_id = $3",
			checkpoint.SucceededAt, checkpoint.FailedAt, checkpoint.SubscriptionId)

		return err
	}

	_, err := dbConnection.ExecContext(monitoringContext,
		"INSERT INTO compaction_checkpoint(subscription_id, succeeded_at, failed_at) VALUES ($1, $2, $3)",
		checkpoint.SubscriptionId, checkpoint.SucceededAt, checkpoint.FailedAt)

	return err
}
