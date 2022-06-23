package integration_test

import (
	"bytes"
	"compress/gzip"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"subscriptions/test/integration/helper"
	"testing"
	"time"
)

func TestSmallFilesAreCompactedForEachDaySinceSubscriptionWasCreatedWhenNoCompactionCheckpoint(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)
	helper.WaitForS3(t)
	helper.ResetAws(t)
	helper.RunTestSetupScript("compact-subscription.sql")

	_, err := http.DefaultClient.Post("http://localhost:8020/cron?cronName=access-log-compaction", "", nil)
	if err != nil {
		t.Fatal("Failed to call cron trigger endpoint", err)
	}

	time.Sleep(time.Second * 5) //todo: replace with exp back off if this works

	gzippedBytes := helper.ReadS3Object(t, "factory-access-log-bucket-int-test", "14fb4f6e-1298-4ca5-989d-00b56a2c6564/2022/06/18/day.gz")

	buf := bytes.NewBuffer(gzippedBytes)
	reader, err := gzip.NewReader(buf)
	if err != nil {
		t.Fatal("Could not read gzip", err)
	}
	defer reader.Close()

	ungzipped, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal("Could not read gzip", err)
	}

	jsonLines := string(ungzipped)

	require.Equal(t,
		`{"Id":  "631eb5f6-0ea3-4f93-a2ff-7bb12a8ddd97", "OccurredAt": 12345, "Product": "Main Product", "Method": "POST", "Path": "/yes/yes", "AndroidId": "f190e8c9-7c62-4d8a-8296-67a100a1f116", "SubscriptionId": "14fb4f6e-1298-4ca5-989d-00b56a2c6564"}
{"Id":  "f47d0838-da93-4bef-a2b1-cbae032b677e", "OccurredAt": 12345, "Product": "Main Product", "Method": "POST", "Path": "/yes/yes", "AndroidId": "f190e8c9-7c62-4d8a-8296-67a100a1f116", "SubscriptionId": "14fb4f6e-1298-4ca5-989d-00b56a2c6564"}
{"Id":  "b8063d5d-b424-4c5a-b410-c5f6f0d4ba2d", "OccurredAt": 12345, "Product": "Main Product", "Method": "POST", "Path": "/yes/yes", "AndroidId": "f190e8c9-7c62-4d8a-8296-67a100a1f116", "SubscriptionId": "14fb4f6e-1298-4ca5-989d-00b56a2c6564"}
{"Id":  "718512dc-2720-4b94-ae1b-ad4e3b5d3404", "OccurredAt": 12345, "Product": "Main Product", "Method": "POST", "Path": "/yes/yes", "AndroidId": "f190e8c9-7c62-4d8a-8296-67a100a1f116", "SubscriptionId": "14fb4f6e-1298-4ca5-989d-00b56a2c6564"}
{"Id":  "972c478c-6f80-4656-bda7-ccc5272898d9", "OccurredAt": 12345, "Product": "Main Product", "Method": "POST", "Path": "/yes/yes", "AndroidId": "f190e8c9-7c62-4d8a-8296-67a100a1f116", "SubscriptionId": "14fb4f6e-1298-4ca5-989d-00b56a2c6564"}
{"Id":  "efe8c75c-4cf5-4aeb-a4d5-0a51c498b78a", "OccurredAt": 12345, "Product": "Main Product", "Method": "POST", "Path": "/yes/yes", "AndroidId": "f190e8c9-7c62-4d8a-8296-67a100a1f116", "SubscriptionId": "14fb4f6e-1298-4ca5-989d-00b56a2c6564"}
{"Id":  "529315c9-6b0c-415e-9763-9bde12fae911", "OccurredAt": 12345, "Product": "Main Product", "Method": "POST", "Path": "/yes/yes", "AndroidId": "f190e8c9-7c62-4d8a-8296-67a100a1f116", "SubscriptionId": "14fb4f6e-1298-4ca5-989d-00b56a2c6564"}
{"Id":  "26c133b9-c7dd-4d72-83d3-daa3f4dabec8", "OccurredAt": 12345, "Product": "Main Product", "Method": "POST", "Path": "/yes/yes", "AndroidId": "f190e8c9-7c62-4d8a-8296-67a100a1f116", "SubscriptionId": "14fb4f6e-1298-4ca5-989d-00b56a2c6564"}
`, jsonLines)
}
