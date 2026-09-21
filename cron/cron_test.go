package cron

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/stretchr/testify/assert"
)

func TestRegisterJob(t *testing.T) {
	// Initialize test data
	testJobName := "testJob"
	testJobFunc := func(s gocron.Scheduler) {}

	// Call the function we want to test
	RegisterJob(testJobName, testJobFunc)

	// Check if the job was registered
	if _, ok := registeredJobs[testJobName]; !ok {
		t.Errorf("job %s was not registered", testJobName)
	}
}

func TestStart(t *testing.T) {
	// Initialize test data
	var test atomic.Int32

	testJobName := "testJob"
	testJobFunc := func(s gocron.Scheduler) {
		_, err := s.NewJob(
			gocron.OneTimeJob(gocron.OneTimeJobStartImmediately()),
			gocron.NewTask(func() {
				test.Add(1)
			}),
		)
		if err != nil {
			t.Errorf("error creating job: %v", err)
		}
	}
	RegisterJob(testJobName, testJobFunc)

	// Call the function we want to test
	Start()

	time.Sleep(1 * time.Second)

	// Check if the job was executed
	assert.Equal(t, int32(1), test.Load(), "testJobFunc was not executed")
}
