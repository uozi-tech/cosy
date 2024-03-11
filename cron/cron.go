package cron

import (
	"github.com/go-co-op/gocron"
	"time"
)

var s *gocron.Scheduler

var registeredJobs map[string]func(*gocron.Scheduler)

func init() {
	s = gocron.NewScheduler(time.Local)
}

func RegisterJob(name string, job func(*gocron.Scheduler)) {
	registeredJobs[name] = job
}

func Start() {
	for _, job := range registeredJobs {
		job(s)
	}
	s.StartAsync()
}
