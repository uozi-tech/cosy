package cosy

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/uozi-tech/cosy/cron"
)

// RegisterCronJob registers a cron job
func RegisterCronJob(name string, job func(gocron.Scheduler)) {
	cron.RegisterJob(name, job)
}
