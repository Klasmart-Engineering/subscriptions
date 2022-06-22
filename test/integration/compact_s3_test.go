package integration_test

import (
	"subscriptions/test/integration/helper"
	"testing"
)

func TestSmallFilesAreCompactedForEachDaySinceSubscriptionWasCreatedWhenNoCompactionCheckpoint(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	//TODO
}

func TestSmallFilesAreCompactedForEachDaySinceLastSuccessWhenExistingCompactionCheckpoint(t *testing.T) {
	helper.ResetDatabase()
	helper.WaitForHealthcheck(t)

	//TODO
}
