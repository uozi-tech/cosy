package logger

import "testing"

func TestCorrelationIndexes(t *testing.T) {
	manager := &SLSManager{}

	apiIndex := manager.createAPILogStoreIndex()
	if key, ok := apiIndex.Keys[FieldCorrelationID]; !ok || key.Type != "text" {
		t.Fatalf("expected API Log correlation index, got %#v", key)
	}

	defaultIndex := manager.createDefaultLogStoreIndex()
	for _, field := range []string{FieldCorrelationID, FieldRequestID, FieldLogType, FieldDBCaller} {
		if key, ok := defaultIndex.Keys[field]; !ok || key.Type != "text" {
			t.Errorf("expected Default Log index for %s, got %#v", field, key)
		}
	}
}
