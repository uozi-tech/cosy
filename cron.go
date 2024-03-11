package cosy

import (
	"git.uozi.org/uozi/cosy/cron"
	"github.com/go-co-op/gocron/v2"
)

// RegisterCronJob registers a cron job
func RegisterCronJob(name string, job func(gocron.Scheduler)) {
	cron.RegisterJob(name, job)
}
