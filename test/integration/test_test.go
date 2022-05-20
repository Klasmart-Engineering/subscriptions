package test_test

import (
	"testing"
	"time"
)

func TestTest(t *testing.T) {
	t.Log("test 1")
	time.Sleep(15000)
	t.Log("test 2")
	time.Sleep(15000)
	t.Log("test 3")
}
