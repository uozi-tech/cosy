package map2struct

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/guregu/null/v6"
	"github.com/jackc/pgtype"
	"github.com/shopspring/decimal"
)

type fuzzNested struct {
	Value int `json:"value"`
}

type fuzzTarget struct {
	String  string            `json:"string"`
	Int     int64             `json:"int"`
	Uint    uint64            `json:"uint"`
	Float   float64           `json:"float"`
	Bool    bool              `json:"bool"`
	Pointer *int              `json:"pointer"`
	Slice   []int             `json:"slice"`
	Map     map[string]string `json:"map"`
	Nested  fuzzNested        `json:"nested"`
	Time    time.Time         `json:"time"`
	TimePtr *time.Time        `json:"time_ptr"`
	Decimal decimal.Decimal   `json:"decimal"`
	Null    null.String       `json:"null"`
	Date    pgtype.Date       `json:"date"`
}

func FuzzWeakDecodeNeverPanics(f *testing.F) {
	seeds := []string{
		`{}`,
		`{"string":1,"int":"2","uint":3,"float":"4.5","bool":"true"}`,
		`{"pointer":null,"slice":["1",2,true],"map":{"key":3},"nested":{"value":"7"}}`,
		`{"string":[],"int":{},"uint":false,"float":[],"bool":{},"slice":{},"nested":[]}`,
		`{"__proto__":{"polluted":true},"STRING":"case-folded","nested":{"value":null}}`,
		`{"time":true,"time_ptr":[],"decimal":false,"null":123,"date":{}}`,
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		var input any
		if err := json.Unmarshal([]byte(raw), &input); err != nil {
			return
		}
		var output fuzzTarget
		_ = WeakDecode(input, &output)
	})
}
