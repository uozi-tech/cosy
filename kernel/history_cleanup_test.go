package kernel

import (
	"sync"
	"testing"
)

func TestHistoryCleanupConcurrentLifecycle(t *testing.T) {
	StopHistoryCleanup()
	t.Cleanup(StopHistoryCleanup)

	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%3 == 0 {
				StopHistoryCleanup()
				return
			}
			StartHistoryCleanup()
		}(i)
	}
	wg.Wait()
	StopHistoryCleanup()

	historyCleanupMu.Lock()
	worker := historyCleanup
	historyCleanupMu.Unlock()
	if worker != nil {
		t.Fatal("expected cleanup worker to be fully stopped")
	}
}
