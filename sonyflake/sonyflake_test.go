package sonyflake

import (
	"github.com/stretchr/testify/assert"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/settings"
	"testing"
	"time"
)

func TestSonyFlake(t *testing.T) {
	logger.Init("debug")

	Init()

	id1 := NextID()

	assert.NotEqual(t, uint64(0), id1)

	id2 := NextID()

	assert.NotEqual(t, id2, id1)

	settings.SonyflakeSettings.StartTime = time.Now()
	settings.SonyflakeSettings.MachineID = 1
	Init()

	id1 = NextID()

	assert.NotEqual(t, uint64(0), id1)

	id2 = NextID()

	assert.NotEqual(t, id2, id1)
}
