package cron

import (
	"github.com/go-co-op/gocron"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRegisterJob(t *testing.T) {
	// Initialize test data
	registeredJobs = make(map[string]func(*gocron.Scheduler))
	testJobName := "testJob"
	testJobFunc := func(s *gocron.Scheduler) {}

	// Call the function we want to test
	RegisterJob(testJobName, testJobFunc)

	// Check if the job was registered
	if _, ok := registeredJobs[testJobName]; !ok {
		t.Errorf("job %s was not registered", testJobName)
	}
}

func TestStart(t *testing.T) {
	// Initialize test data
	test := 0

	registeredJobs = make(map[string]func(*gocron.Scheduler))
	testJobName := "testJob"
	testJobFunc := func(s *gocron.Scheduler) {
		_, err := s.Every(1).Second().Do(func() {
			test = 1
		})
		if err != nil {
			t.Errorf("error creating job: %v", err)
		}
	}
	RegisterJob(testJobName, testJobFunc)

	// Call the function we want to test
	Start()

	time.Sleep(2 * time.Second)

	// Check if the job was executed
	assert.Equal(t, 1, test, "testJobFunc was not executed")
}
