package sonyflake

import (
	"git.uozi.org/uozi/cosy/logger"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSonyFlake(t *testing.T) {
	logger.Init("debug")

	Init()

	id1 := NextID()

	assert.NotEqual(t, uint64(0), id1)

	id2 := NextID()

	assert.NotEqual(t, id2, id1)
}
