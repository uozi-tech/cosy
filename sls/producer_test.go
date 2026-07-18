package sls

import (
	"testing"
	"time"
)

func TestProducerSendLogAppliesBackpressure(t *testing.T) {
	producer := &Producer{
		ch:   make(chan logEntry, 1),
		quit: make(chan struct{}),
	}
	if err := producer.SendLog("project", "store", "", "", &Log{}); err != nil {
		t.Fatalf("enqueue first log: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- producer.SendLog("project", "store", "", "", &Log{})
	}()
	select {
	case err := <-done:
		t.Fatalf("expected full queue to block, returned %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	<-producer.ch
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("enqueue after drain: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("blocked sender did not resume after queue drained")
	}
}

func TestProducerSendLogRejectsClosedProducer(t *testing.T) {
	producer := &Producer{
		ch:   make(chan logEntry, 1),
		quit: make(chan struct{}),
	}
	producer.SafeClose()
	producer.SafeClose()
	if err := producer.SendLog("project", "store", "", "", &Log{}); err == nil {
		t.Fatal("expected closed producer error")
	}
}
