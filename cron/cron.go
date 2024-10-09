package cron

import (
    "github.com/go-co-op/gocron/v2"
    "github.com/uozi-tech/cosy/logger"
)

var s gocron.Scheduler

var registeredJobs map[string]func(gocron.Scheduler)

func init() {
    var err error
    s, err = gocron.NewScheduler()
    if err != nil {
        logger.Fatal(err)
    }
    registeredJobs = make(map[string]func(gocron.Scheduler))
}

// RegisterJob registers a job to be run by the scheduler
func RegisterJob(name string, job func(gocron.Scheduler)) {
    registeredJobs[name] = job
}

// Start starts the scheduler
func Start() {
    for _, job := range registeredJobs {
        job(s)
    }
    s.Start()
}
