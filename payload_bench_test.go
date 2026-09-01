package cosy

import (
	"encoding/json"
	"testing"
)

// A typical 10-field write payload, mirroring the model used in
// PERF_REFACTOR_PLAN.md benchmarks.
var benchPayloadBody = []byte(`{
	"school_id": "0281876",
	"name": "张三",
	"age": 20,
	"title": "助理教授",
	"bio": "大数据与人工智能方向",
	"college": "大数据与互联网学院",
	"balance": "1234.56",
	"status": 1,
	"employed_at": "2024-03-13T11:22:44.405374+08:00",
	"tags": ["a", "b", "c"]
}`)

func BenchmarkStdJSONToMap(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		m := make(map[string]any)
		if err := json.Unmarshal(benchPayloadBody, &m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPipelineJSONToMap(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		m := make(map[string]any)
		if err := decodeJSON(benchPayloadBody, &m); err != nil {
			b.Fatal(err)
		}
	}
}
