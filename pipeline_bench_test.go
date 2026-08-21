package cosy

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guregu/null/v6"
	"github.com/shopspring/decimal"
	"github.com/uozi-tech/cosy/internal/rulecheck"
	"github.com/uozi-tech/cosy/map2struct"
)

// benchPipelineModel mirrors the "typical write model" used throughout
// PERF_REFACTOR_PLAN.md: 10 fields including time.Time, decimal.Decimal,
// null.String and a slice.
type benchPipelineModel struct {
	SchoolID   string          `json:"school_id"`
	Name       string          `json:"name"`
	Age        int             `json:"age"`
	Title      string          `json:"title"`
	Bio        null.String     `json:"bio"`
	College    string          `json:"college"`
	Balance    decimal.Decimal `json:"balance"`
	Status     int             `json:"status"`
	EmployedAt time.Time       `json:"employed_at"`
	Tags       []string        `json:"tags"`
}

var benchPipelineRules = gin.H{
	"school_id":   "required,max=20",
	"name":        "required,safety_text,max=50",
	"age":         "min=0,max=150",
	"title":       "omitempty,max=50",
	"bio":         "omitempty,max=200",
	"college":     "omitempty,max=50",
	"balance":     "omitempty",
	"status":      "min=0,max=2",
	"employed_at": "omitempty",
	"tags":        "omitempty",
}

func benchPipelineParse(b *testing.B) gin.H {
	b.Helper()
	payload := make(gin.H)
	if err := decodeJSON(benchPayloadBody, &payload); err != nil {
		b.Fatal(err)
	}
	return payload
}

func benchPipelineValidate(b *testing.B, payload gin.H) {
	b.Helper()
	if errs := rulecheck.ValidateMap(v, payload, benchPipelineRules); len(errs) != 0 {
		b.Fatal(errs)
	}
}

// BenchmarkPipelineValidate: stage 2, map -> rule check, on a pre-parsed map.
func BenchmarkPipelineValidate(b *testing.B) {
	payload := benchPipelineParse(b)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		benchPipelineValidate(b, payload)
	}
}

// BenchmarkPipelineDecode: stage 3, map -> struct, on a pre-parsed map.
func BenchmarkPipelineDecode(b *testing.B) {
	payload := benchPipelineParse(b)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var model benchPipelineModel
		if err := map2struct.WeakDecode(payload, &model); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPipelineEndToEnd: bytes -> map -> validate -> struct, the framework
// side of one Create/Update request.
func BenchmarkPipelineEndToEnd(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		payload := make(gin.H)
		if err := decodeJSON(benchPayloadBody, &payload); err != nil {
			b.Fatal(err)
		}
		benchPipelineValidate(b, payload)
		var model benchPipelineModel
		if err := map2struct.WeakDecode(payload, &model); err != nil {
			b.Fatal(err)
		}
	}
}
