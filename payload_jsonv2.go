//go:build go1.27

package cosy

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
)

// jsonv2Options keeps encoding/json/v2 on the same semantics as the sonic path
// and encoding/json v1: duplicate keys are last-wins and invalid UTF-8 is
// coerced to U+FFFD. Dropping these two options switches the pipeline to v2's
// strict defaults (both cases rejected).
var jsonv2Options = jsonv2.JoinOptions(
	jsontext.AllowDuplicateNames(true),
	jsontext.AllowInvalidUTF8(true),
)

type jsonv2Decoder struct{}

func (jsonv2Decoder) Unmarshal(buf []byte, val any) error {
	return jsonv2.Unmarshal(buf, val, jsonv2Options)
}

// jsonDecoder decodes request bodies with encoding/json/v2. Strings are always
// copied (I3) and raw control characters are rejected.
var jsonDecoder jsonv2Decoder
