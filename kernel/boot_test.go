package kernel

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var a = 1

func TestBoot(t *testing.T) {
	ctx := context.Background()
	RegisterInitFunc(func() {
		a = 2
	})
	RegisterGoroutine(func(ctx context.Context) {
		a = 3
	})
	Boot(ctx)
	time.Sleep(1 * time.Second)

	assert.Equal(t, 3, a, "a should be 3")
}
