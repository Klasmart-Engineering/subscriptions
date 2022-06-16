package cron

import (
	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
	db "subscriptions/src/database"
	"subscriptions/src/monitoring"
	"time"
)

func StartCronJobs() {
	monitoring.GlobalContext.Info("Scheduling cron jobs")
	scheduler := gocron.NewScheduler(time.UTC)
	_, err := scheduler.Every(1).Day().At("00:20").Do(AttemptToLockThenDo("access-log-compaction", CompactionCron))
	if err != nil {
		monitoring.GlobalContext.Fatal("Unable to schedule access log compaction", zap.Error(err))
	}

	scheduler.StartAsync()
}

func AttemptToLockThenDo(cronName string, action func()) func() {
	return func() {
		gotLock := db.AttemptToGetLock(cronName)

		if gotLock {
			monitoring.GlobalContext.Info("Got lock for cron " + cronName + ".  Performing task")
			action()
			return
		}

		monitoring.GlobalContext.Info("Could not get lock for cron " + cronName)
	}
}

func CompactionCron() {
	monitoring.GlobalContext.Info("Todo: compact S3 entries into day entry")
}
