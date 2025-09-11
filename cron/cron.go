package cron

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/uozi-tech/cosy/logger"
)

var s gocron.Scheduler

var registeredJobs = make(map[string]func(gocron.Scheduler))

func init() {
	var err error
	s, err = gocron.NewScheduler()
	if err != nil {
		logger.Fatal(err)
	}
}

// RegisterJob registers a job to be run by the scheduler
func RegisterJob(name string, job func(gocron.Scheduler)) {
	registeredJobs[name] = job
}

// Start start the scheduler
func Start() {
	var err error
	s, err = gocron.NewScheduler()
	if err != nil {
		logger.Fatal(err)
	}
	for _, job := range registeredJobs {
		job(s)
	}
	s.Start()
}

// Stop stop the scheduler
func Stop() {
	s.Shutdown()
}
