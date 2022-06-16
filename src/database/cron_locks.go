package db

import (
	"database/sql"
	"go.uber.org/zap"
	"os"
	"subscriptions/src/monitoring"
)

func AttemptToGetLock(cronName string) bool {
	transaction, err := dbConnection.BeginTxx(monitoring.GlobalContext, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	if err != nil {
		monitoring.GlobalContext.Error("Could not start transaction to get lock for cron job: "+cronName, zap.Error(err))
		return false
	}

	var selectedName *string
	err = transaction.Get(&selectedName, "SELECT name FROM cron_job_lock WHERE name = $1 AND locked_until < NOW() AT TIME ZONE 'UTC' FOR UPDATE NOWAIT", cronName)
	if err != nil {
		transaction.Commit()
		return false
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "Unknown"
	}

	_, err = transaction.Exec("UPDATE cron_job_lock SET locked_by = $1, locked_until = NOW() AT TIME ZONE 'UTC' + INTERVAL '24 HOUR' WHERE name = $2",
		hostname, cronName)

	if err != nil {
		transaction.Commit()
		return false
	}

	err = transaction.Commit()

	if err != nil {
		monitoring.GlobalContext.Error("Could not end transaction on cron lock table", zap.Error(err))
		return false
	}

	return true
}
