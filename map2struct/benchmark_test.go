package map2struct

import (
	"testing"
	"time"

	"github.com/guregu/null/v6"
	"github.com/jackc/pgtype"
	"github.com/shopspring/decimal"
)

type benchmarkProfile struct {
	Bio     null.String `json:"bio"`
	Country string      `json:"country"`
}

type benchmarkModel struct {
	ID        uint64           `json:"id"`
	Name      string           `json:"name"`
	Age       int              `json:"age"`
	Active    bool             `json:"active"`
	Score     float64          `json:"score"`
	Balance   decimal.Decimal  `json:"balance"`
	CreatedAt time.Time        `json:"created_at"`
	Birthday  *pgtype.Date     `json:"birthday"`
	Tags      []string         `json:"tags"`
	Profile   benchmarkProfile `json:"profile"`
}

var benchmarkPayload = map[string]any{
	"id":         float64(42),
	"name":       "Ada",
	"age":        "37",
	"active":     "true",
	"score":      float64(98.5),
	"balance":    "12345.6789",
	"created_at": "2026-07-03T12:34:56Z",
	"birthday":   "1989-04-21",
	"tags":       []any{"admin", "author", "reviewer"},
	"profile": map[string]any{
		"bio":     "compiler engineer",
		"country": "CN",
	},
}

func BenchmarkWeakDecode(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var output benchmarkModel
		if err := WeakDecode(benchmarkPayload, &output); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWeakDecodeWarmParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var output benchmarkModel
			if err := WeakDecode(benchmarkPayload, &output); err != nil {
				b.Fatal(err)
			}
		}
	})
}
