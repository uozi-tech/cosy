package settings

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerShutdownTimeout(t *testing.T) {
	assert := assert.New(t)

	t.Run("Unset falls back to the default", func(t *testing.T) {
		assert.Equal(DefaultShutdownTimeout, (&Server{}).ShutdownTimeout())
		assert.Equal(DefaultShutdownTimeout, (&Server{ShutdownTimeoutSeconds: 0}).ShutdownTimeout())
	})

	t.Run("Default is longer than the old hardcoded 5s", func(t *testing.T) {
		assert.Greater(DefaultShutdownTimeout, 5*time.Second)
	})

	t.Run("Default leaves room inside the 30s Kubernetes grace period", func(t *testing.T) {
		assert.Less(DefaultShutdownTimeout, 30*time.Second)
	})

	t.Run("A positive value is taken as seconds", func(t *testing.T) {
		assert.Equal(1*time.Second, (&Server{ShutdownTimeoutSeconds: 1}).ShutdownTimeout())
		assert.Equal(600*time.Second, (&Server{ShutdownTimeoutSeconds: 600}).ShutdownTimeout())
	})

	t.Run("A negative value is not unbounded", func(t *testing.T) {
		assert.Equal(DefaultShutdownTimeout, (&Server{ShutdownTimeoutSeconds: -1}).ShutdownTimeout())
	})

	t.Run("ServerSettings resolves a usable budget", func(t *testing.T) {
		prev := ServerSettings.ShutdownTimeoutSeconds
		t.Cleanup(func() { ServerSettings.ShutdownTimeoutSeconds = prev })

		ServerSettings.ShutdownTimeoutSeconds = 0
		assert.Equal(DefaultShutdownTimeout, ServerSettings.ShutdownTimeout())

		ServerSettings.ShutdownTimeoutSeconds = 20
		assert.Equal(20*time.Second, ServerSettings.ShutdownTimeout())
	})
}
