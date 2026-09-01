package cosy

import (
	"encoding/json"
	jsonv2 "encoding/json/v2"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type probeDecoder struct {
	name   string
	decode func([]byte, *map[string]any) error
}

var probeDecoders = []probeDecoder{
	{"std-v1", func(b []byte, m *map[string]any) error { return json.Unmarshal(b, m) }},
	{"jsonv2", func(b []byte, m *map[string]any) error { return jsonv2.Unmarshal(b, m) }},
	{"cosy-pipeline", func(b []byte, m *map[string]any) error { return decodeJSON(b, m) }},
}

// TestJSONDecoderSemanticsMatrix records how encoding/json v1, json/v2 and the
// pipeline decoder treat malformed and boundary inputs. Per-decoder behaviour
// is logged, while every case is required to return without panicking.
func TestJSONDecoderSemanticsMatrix(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"dup-keys", `{"a":1,"a":2}`},
		{"invalid-utf8", "{\"a\":\"\xff\"}"},
		{"raw-control-char", "{\"a\":\"x\x01y\"}"},
		{"huge-number", `{"a":1e309}`},
		{"int64-overflow", `{"a":18446744073709551616}`},
		{"depth-5000", `{"a":` + strings.Repeat("[", 5000) + strings.Repeat("]", 5000) + `}`},
		{"depth-20000", `{"a":` + strings.Repeat("[", 20000) + strings.Repeat("]", 20000) + `}`},
		{"empty", ``},
		{"null", `null`},
		{"array-root", `[1,2]`},
		{"trailing-garbage", `{"a":1} x`},
	}
	for _, tc := range cases {
		for _, d := range probeDecoders {
			t.Run(tc.name+"/"+d.name, func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panic: %v", r)
					}
				}()
				m := make(map[string]any)
				err := d.decode([]byte(tc.body), &m)
				if err != nil {
					msg := err.Error()
					if i := strings.IndexByte(msg, '\n'); i > 0 {
						msg = msg[:i]
					}
					t.Logf("err=%s", msg)
					return
				}
				v := fmt.Sprintf("%#v", m["a"])
				if len(v) > 60 {
					v = v[:60] + "..."
				}
				t.Logf("ok value=%s", v)
			})
		}
	}
}

func TestJSONDecoderCopiesStrings(t *testing.T) {
	raw := []byte(`{"name":"copystring-check"}`)
	var decoded map[string]any
	require.NoError(t, decodeJSON(raw, &decoded))

	for i := range raw {
		raw[i] = 'z'
	}
	assert.Equal(t, "copystring-check", decoded["name"])
}

func TestJSONDecoderRejectsRawControlChars(t *testing.T) {
	var decoded map[string]any
	err := decodeJSON([]byte("{\"a\":\"x\x01y\"}"), &decoded)
	assert.Error(t, err)
}
