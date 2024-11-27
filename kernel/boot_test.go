package kernel

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var a = 1

func TestBoot(t *testing.T) {
	RegisterInitFunc(func() {
		a = 2
	})
	RegisterGoroutine(func() {
		a = 3
	})
	Boot()
	time.Sleep(1 * time.Second)

	assert.Equal(t, 3, a, "a should be 3")
}
