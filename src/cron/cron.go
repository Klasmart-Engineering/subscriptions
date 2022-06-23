package cron

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-co-op/gocron"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"subscriptions/src/aws"
	"subscriptions/src/config"
	db "subscriptions/src/database"
	"subscriptions/src/models"
	"subscriptions/src/monitoring"
	"subscriptions/src/utils"
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
func ForceCronJob(c echo.Context) error {
	switch c.QueryParam("cronName") {
	case "access-log-compaction":
		CompactionCron()
		c.NoContent(200)
	default:
		c.NoContent(404)
	}

	return nil
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

const subscriptionsPageSize = 50

func CompactionCron() {
	subscriptionIdsOffset := 0
	page, err := db.GetSubscriptionsPage(monitoring.GlobalContext, subscriptionsPageSize, subscriptionIdsOffset)
	if err != nil {
		monitoring.GlobalContext.Error("Could not get page of Subscription Ids when attempting to compact into day "+
			"objects", zap.Error(err))
		return
	}

	for _, subscription := range page {
		go processSubscription(subscription)
	}
}

func processSubscription(subscription models.Subscription) {
	monitoring.GlobalContext.Info("Starting s3 compact", zap.String("subscriptionId", subscription.Id.String()))
	checkpointExists, checkpoint, err := db.GetCompactionCheckpoint(monitoring.GlobalContext, subscription.Id)
	if err != nil {
		monitoring.GlobalContext.Error("Could not get compaction checkpoint ",
			zap.Error(err), zap.String("subscriptionId", subscription.Id.String()))
		return
	}

	var currentDay time.Time
	var end = utils.ToDay(time.Now())
	if !checkpointExists || checkpoint.SucceededAt == nil {
		currentDay = utils.ToDay(subscription.CreatedAt)
		checkpoint = models.CompactionCheckpoint{
			SubscriptionId: subscription.Id,
			SucceededAt:    nil,
			FailedAt:       nil,
		}
	} else {
		currentDay = utils.ToDay(*checkpoint.SucceededAt)
	}

	for currentDay.Before(end) {
		monitoring.GlobalContext.Info("Starting s3 compact day", zap.String("subscriptionId", subscription.Id.String()), zap.Time("day", currentDay))
		if err = processSubscriptionDay(subscription, checkpoint, currentDay); err != nil {
			return
		}
		monitoring.GlobalContext.Info("Finished s3 compact day", zap.String("subscriptionId", subscription.Id.String()), zap.Time("day", currentDay))

		currentDay = currentDay.Add(time.Hour * 24)
	}
}

func processSubscriptionDay(subscription models.Subscription, checkpoint models.CompactionCheckpoint, day time.Time) error {
	var continuationToken *string
	onlyDelete := false
	anythingWritten := false

	dayFileName := getTmpFileName(subscription, day)

	if utils.FileExists(dayFileName) {
		err := os.Remove(dayFileName)
		if err != nil {
			monitoring.GlobalContext.Error("Could not remove day file", zap.Error(err))
			return err
		}
	}

	dayFile, err := os.Create(dayFileName)
	if err != nil {
		monitoring.GlobalContext.Error("Could not create day file", zap.Error(err))
		return err
	}

continuationLoop:
	for {
		response, err := aws.S3Client.ListObjectsV2(monitoring.GlobalContext, &s3.ListObjectsV2Input{
			Bucket:            &config.GetConfig().BucketConfig.AccessLogBucket,
			ContinuationToken: continuationToken,
			MaxKeys:           50,
			Prefix:            utils.StringPtr(fmt.Sprintf("%s/%s", subscription.Id.String(), day.Format("2006/01/02"))),
		})
		if err != nil {
			monitoring.GlobalContext.Error("Could not list objects when attempting to compact into day "+
				"object for Subscription", zap.Error(err), zap.String("subscriptionId", subscription.Id.String()), zap.Time("day", day))
			attemptToMarkLastFailure(checkpoint)
			return err
		}

		for _, objectInfo := range response.Contents {
			if strings.Contains(*objectInfo.Key, "/day") {
				onlyDelete = true
				break continuationLoop
			}

			object, err := aws.S3Client.GetObject(monitoring.GlobalContext, &s3.GetObjectInput{
				Bucket: utils.StringPtr(config.GetConfig().BucketConfig.AccessLogBucket),
				Key:    objectInfo.Key,
			})
			if err != nil {
				monitoring.GlobalContext.Error("Could not get object when attempting to compact into day "+
					"object for Subscription", zap.Error(err), zap.String("subscriptionId", subscription.Id.String()), zap.Time("day", day))
				attemptToMarkLastFailure(checkpoint)
				return err
			}

			all, err := io.ReadAll(object.Body)
			if err != nil {
				monitoring.GlobalContext.Error("Could not read object when attempting to compact into day "+
					"object for Subscription", zap.Error(err), zap.String("subscriptionId", subscription.Id.String()), zap.Time("day", day))
				attemptToMarkLastFailure(checkpoint)
				return err
			}

			gz, err := gzip.NewReader(bytes.NewBuffer(all))
			if err != nil {
				monitoring.GlobalContext.Error("Could not create gzip reader for Subscription",
					zap.Error(err), zap.String("subscriptionId", subscription.Id.String()), zap.Time("day", day))
				attemptToMarkLastFailure(checkpoint)
				return err
			}

			ungzipped, err := ioutil.ReadAll(gz)
			gz.Close()
			if err != nil {
				monitoring.GlobalContext.Error("Could not un-gzip object for Subscription",
					zap.Error(err), zap.String("subscriptionId", subscription.Id.String()), zap.Time("day", day))
				attemptToMarkLastFailure(checkpoint)
				return err
			}

			_, err = dayFile.Write(append(ungzipped, "\n"...))
			if err != nil {
				monitoring.GlobalContext.Error("Could not write to day file when attempting to compact into day "+
					"object for Subscription", zap.Error(err), zap.String("subscriptionId", subscription.Id.String()), zap.Time("day", day))
				return err
			}

			anythingWritten = true
		}

		continuationToken = response.ContinuationToken

		if continuationToken == nil {
			break
		}
	}

	err = dayFile.Close()
	if err != nil {
		monitoring.GlobalContext.Error("Could not close day file when attempting to compact into day "+
			"object for Subscription", zap.Error(err), zap.String("subscriptionId", subscription.Id.String()), zap.Time("day", day))
		return err
	}

	if !onlyDelete && anythingWritten {
		if err := writeDayFileFromTmp(subscription, day); err != nil {
			monitoring.GlobalContext.Error("Unable to write day file",
				zap.Error(err), zap.String("subscriptionId", checkpoint.SubscriptionId.String()), zap.Time("succeededAt", *checkpoint.SucceededAt))

			return err
		}
	}

	if anythingWritten {
		var deleteContinuationToken *string
		for {
			response, err := aws.S3Client.ListObjectsV2(monitoring.GlobalContext, &s3.ListObjectsV2Input{
				Bucket:            &config.GetConfig().BucketConfig.AccessLogBucket,
				ContinuationToken: deleteContinuationToken,
				MaxKeys:           50,
				Prefix:            utils.StringPtr(fmt.Sprintf("%s/%s", subscription.Id.String(), day.Format("2006/01/02"))),
			})
			if err != nil {
				monitoring.GlobalContext.Error("Could not list objects when attempting to delete small "+
					"objects for Subscription", zap.Error(err), zap.String("subscriptionId", subscription.Id.String()), zap.Time("day", day))
				attemptToMarkLastFailure(checkpoint)
				return err
			}

			for _, objectInfo := range response.Contents {
				if strings.Contains(*objectInfo.Key, "/day") {
					continue
				}

				_, err := aws.S3Client.DeleteObject(monitoring.GlobalContext, &s3.DeleteObjectInput{
					Bucket: &config.GetConfig().BucketConfig.AccessLogBucket,
					Key:    objectInfo.Key,
				})
				if err != nil {
					monitoring.GlobalContext.Error("Could not delete small object when attempting to delete small "+
						"objects for Subscription", zap.Error(err), zap.String("subscriptionId", subscription.Id.String()),
						zap.Time("day", day), zap.String("key", *objectInfo.Key))
					return err
				}
			}

			deleteContinuationToken = response.ContinuationToken

			if deleteContinuationToken == nil {
				break
			}
		}
	}

	if err := markLastSuccess(checkpoint); err != nil {
		monitoring.GlobalContext.Error("Unable to mark last success on compaction checkpoint",
			zap.Error(err), zap.String("subscriptionId", checkpoint.SubscriptionId.String()), zap.Time("succeededAt", *checkpoint.SucceededAt))

		return err
	}

	return nil
}

func writeDayFileFromTmp(subscription models.Subscription, day time.Time) error {
	file, err := os.Open(getTmpFileName(subscription, day))
	if err != nil {
		monitoring.GlobalContext.Error("Could not open tmp day file to write to S3", zap.Error(err))
		return err
	}

	handle, err := os.Create(getTmpFileName(subscription, day) + ".gz")
	if err != nil {
		monitoring.GlobalContext.Error("Could not open tmp day gz file to write to S3", zap.Error(err))
		return err
	}

	zipWriter, err := gzip.NewWriterLevel(handle, 9)
	if err != nil {
		monitoring.GlobalContext.Error("Could not create tmp day gz writer to write to S3", zap.Error(err))
		return err
	}

	for {
		buf := make([]byte, 0, 4*1024)
		n, err := file.Read(buf[:cap(buf)])
		buf = buf[:n]
		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}

		}

		if err != nil && err != io.EOF {
			monitoring.GlobalContext.Error("Could not read day file into gz writer to write to S3", zap.Error(err))
			return err
		}

		_, err = zipWriter.Write(buf)
		if err != nil && err != io.EOF {
			monitoring.GlobalContext.Error("Could not write to zip writer", zap.Error(err))
			return err
		}
	}

	zipWriter.Close()
	handle.Close()

	dayFileUploader := manager.NewUploader(aws.S3Client, func(u *manager.Uploader) {
		u.PartSize = 0
		u.Concurrency = 1
		u.LeavePartsOnError = false
	})

	gzippedFile, err := os.Open(getTmpFileName(subscription, day) + ".gz")

	_, err = dayFileUploader.Upload(monitoring.GlobalContext, &s3.PutObjectInput{
		Bucket: utils.StringPtr(config.GetConfig().BucketConfig.AccessLogBucket),
		Key:    utils.StringPtr(fmt.Sprintf("%s/%s/day.gz", subscription.Id.String(), day.Format("2006/01/02"))),
		Body:   gzippedFile,
	})

	gzippedFile.Close()

	removeErr := os.Remove(getTmpFileName(subscription, day))
	if removeErr != nil {
		monitoring.GlobalContext.Error("Could not delete tmp day file", zap.String("subscriptionId",
			subscription.Id.String()), zap.Error(removeErr))
	}

	return err
}

func getTmpFileName(subscription models.Subscription, day time.Time) string {
	return fmt.Sprintf("/tmp/dayfiles/%s-%s-day", subscription.Id.String(), day.Format("2006-01-02"))
}

func attemptToMarkLastFailure(checkpoint models.CompactionCheckpoint) {
	checkpoint.FailedAt = utils.TimePtr(time.Now())
	err := db.UpsertCompactionCheckpoint(monitoring.GlobalContext, checkpoint)
	if err != nil {
		monitoring.GlobalContext.Error("Unable to mark last failure on compaction checkpoint",
			zap.Error(err), zap.String("subscriptionId", checkpoint.SubscriptionId.String()), zap.Time("failedAt", *checkpoint.FailedAt))
	}
}

func markLastSuccess(checkpoint models.CompactionCheckpoint) error {
	checkpoint.SucceededAt = utils.TimePtr(time.Now())
	return db.UpsertCompactionCheckpoint(monitoring.GlobalContext, checkpoint)
}
