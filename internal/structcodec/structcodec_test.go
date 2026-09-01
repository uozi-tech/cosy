package structcodec

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeConcurrentPlanCache(t *testing.T) {
	type nested struct {
		Value int `json:"value"`
	}
	type target struct {
		Name   string `json:"name"`
		Nested nested `json:"nested"`
	}

	const workers = 64
	var waitGroup sync.WaitGroup
	errors := make(chan error, workers)
	for range workers {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			var got target
			if err := Decode(map[string]any{
				"name":   42,
				"nested": map[string]any{"value": "7"},
			}, &got); err != nil {
				errors <- err
				return
			}
			if got.Name != "42" || got.Nested.Value != 7 {
				errors <- fmt.Errorf("unexpected decode result: %#v", got)
			}
		}()
	}
	waitGroup.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
}

func TestDecodeRejectsInvalidOutput(t *testing.T) {
	assert.EqualError(t, Decode(map[string]any{}, struct{}{}), "result must be a pointer")
	var output *struct{}
	assert.EqualError(t, Decode(map[string]any{}, output), "result must be addressable (a pointer)")
}
