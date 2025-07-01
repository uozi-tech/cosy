package geoip

import (
	"github.com/stretchr/testify/assert"
	"github.com/uozi-tech/cosy/logger"
	"testing"
)

func TestParseIP(t *testing.T) {
	logger.Init("debug")

	assert.Equal(t, "US", ParseIP("8.8.8.8"))
}
