package cron

import (
	"git.uozi.org/uozi/cosy/logger"
	"github.com/go-co-op/gocron/v2"
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

func RegisterJob(name string, job func(gocron.Scheduler)) {
	registeredJobs[name] = job
}

func Start() {
	for _, job := range registeredJobs {
		job(s)
	}
	s.Start()
}
