package helper

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/cenkalti/backoff/v4"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"io"
	"log"
	"net/http"
	"os"
	"subscriptions/src/utils"
	db "subscriptions/test/integration/db"
	"testing"
	"time"
)

var s3Client *s3.Client

var dropStatements = readFile("../../database/drop-all-tables.sql")

func readFile(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		log.Panicf("Could not read %s", file)
	}
	return string(content)
}

func ResetDatabase() {
	initIfNeeded()

	execOrPanic(dropStatements)
	driver, err := postgres.WithInstance(db.DbConnection, &postgres.Config{})
	if err != nil {
		log.Panicf("Could not create migration driver: %s", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../database/migrations",
		"postgres", driver)

	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No migrations to run, up to date")
		} else {
			log.Fatalf("Could not migrate database: %s", err)
		}
	}
}

func RunTestSetupScript(fileName string) {
	initIfNeeded()

	execOrPanic(readFile("./test-sql/" + fileName))
}

func GetDatabaseConnection() *sql.DB {
	initIfNeeded()

	return db.DbConnection
}

func ExactlyOneRowMatches(query string) error {
	var rowCount int
	if err := GetDatabaseConnection().QueryRow(query).Scan(&rowCount); err != nil {
		if err == sql.ErrNoRows {
			panic(err)
		}
	}

	if rowCount != 1 {
		return fmt.Errorf("query \"%s\" returned %d rows instead of 1", query, rowCount)
	}

	return nil
}

func AssertExactlyOneRowMatchesWithBackoff(t *testing.T, query string) {
	var check = func() error {
		return ExactlyOneRowMatches(query)
	}

	err := backoff.Retry(check, &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.2,
		MaxInterval:         1 * time.Second,
		MaxElapsedTime:      5 * time.Second,
		Stop:                -1,
		Clock:               backoff.SystemClock,
	})

	if err != nil {
		t.Fatal(err)
	}
}

func execOrPanic(statement string) {
	_, err := db.DbConnection.Exec(statement)

	if err != nil {
		log.Panicf("Could not execute database statement %s: %s", statement, err)
	}
}

func initIfNeeded() {
	if db.DbConnection == nil {
		err := db.Initialize(
			"postgres",
			"integration-test-pa55word!",
			"subscriptions",
			"localhost",
			1334)

		if err != nil {
			log.Panicf("Could not connect to database to reset for integration tests: %s", err)
		}
	}
}

func WaitForHealthcheck(t *testing.T) {
	var check = func() error {
		_, err := http.Get("http://localhost:8020/healthcheck")
		if err != nil {
			return err
		}

		return nil
	}

	err := backoff.Retry(check, &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.2,
		MaxInterval:         1 * time.Second,
		MaxElapsedTime:      30 * time.Second,
		Stop:                -1,
		Clock:               backoff.SystemClock,
	})

	if err != nil {
		t.Fatal(err)
	}
}

func setupAWS() {
	if s3Client != nil {
		return
	}

	creds := credentials.NewStaticCredentialsProvider(
		"test",
		"test",
		"")

	cfg := aws.Config{
		Credentials: creds,
		Region:      "eu-west-1",
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           "http://localhost:4568",
				SigningRegion: region,
			}, nil
		}),
		EndpointResolver: aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			//Despite being deprecated, it seems this is actually still used sometimes
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           "http://localhost:4568",
				SigningRegion: region,
			}, nil
		}),
	}

	s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
}

func WaitForS3(t *testing.T) {
	setupAWS()

	var check = func() (returnedError error) {
		defer func() {
			err := recover()
			if err != nil {
				returnedError = err.(error)
			}
		}()

		_, err := s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
		if err != nil {
			return err
		}

		return nil
	}

	err := backoff.Retry(check, &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.2,
		MaxInterval:         20 * time.Second,
		MaxElapsedTime:      120 * time.Second,
		Stop:                -1,
		Clock:               backoff.SystemClock,
	})

	if err != nil {
		t.Fatalf("Timed out waiting for S3: %s", err)
	}
}

func ResetAws(t *testing.T) {
	buckets, err := s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		t.Fatal("Unable to list buckets", err)
	}

	found := false
	for _, bucket := range buckets.Buckets {
		if *bucket.Name == "factory-access-log-bucket-int-test" {
			found = true
		}
	}

	if !found {
		_, err := s3Client.CreateBucket(context.Background(), &s3.CreateBucketInput{
			Bucket: utils.StringPtr("factory-access-log-bucket-int-test"),
			ACL:    "public-read-write",
			CreateBucketConfiguration: &types.CreateBucketConfiguration{
				LocationConstraint: types.BucketLocationConstraintEuWest1,
			},
		})
		if err != nil {
			t.Fatal("Could not create bucket", err)
		}
	}

	objects, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: utils.StringPtr("factory-access-log-bucket-int-test"),
	})
	if err != nil {
		t.Fatal("Could not list objects to delete", err)
	}

	for _, content := range objects.Contents {
		_, err := s3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
			Bucket: utils.StringPtr("factory-access-log-bucket-int-test"),
			Key:    content.Key,
		})
		if err != nil {
			t.Fatal("Could not delete object", err)
		}
	}

	putFileS3(t, "./small-files/1.gz", "14fb4f6e-1298-4ca5-989d-00b56a2c6564/2022/06/17/1.gz")
	putFileS3(t, "./small-files/2.gz", "14fb4f6e-1298-4ca5-989d-00b56a2c6564/2022/06/18/2.gz")
	putFileS3(t, "./small-files/3.gz", "14fb4f6e-1298-4ca5-989d-00b56a2c6564/2022/06/18/3.gz")
	putFileS3(t, "./small-files/4.gz", "14fb4f6e-1298-4ca5-989d-00b56a2c6564/2022/06/19/4.gz")
}

func putFileS3(t *testing.T, fileName string, s3Location string) {
	file, err := os.Open(fileName)
	if err != nil {
		t.Fatal("Could not open file", err)
	}

	dayFileUploader := manager.NewUploader(s3Client, func(u *manager.Uploader) {
		u.PartSize = 0
		u.Concurrency = 1
		u.LeavePartsOnError = false
	})

	_, err = dayFileUploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket: utils.StringPtr("factory-access-log-bucket-int-test"),
		Key:    &s3Location,
		Body:   file,
	})
}

func ReadS3Object(t *testing.T, bucket string, objectName string) []byte {
	var all []byte
	var check = func() error {
		object, err := s3Client.GetObject(context.Background(), &s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &objectName,
		})
		if err != nil {
			return err
		}

		all, err = io.ReadAll(object.Body)

		return err
	}

	err := backoff.Retry(check, &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.2,
		MaxInterval:         20 * time.Second,
		MaxElapsedTime:      120 * time.Second,
		Stop:                -1,
		Clock:               backoff.SystemClock,
	})
	if err != nil {
		t.Fatal("Unable to read object", err)
	}

	return all
}
