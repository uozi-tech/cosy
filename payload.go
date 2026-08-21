package cosy

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/settings"
)

// defaultPayloadMaxBytes caps JSON request bodies parsed by the CRUD pipeline
// and BindAndValid when settings.ServerSettings.PayloadMaxBytes is 0 (F2).
const defaultPayloadMaxBytes int64 = 10 << 20 // 10 MiB

// jsonv2Options keeps encoding/json/v2 on the semantics the pipeline had under
// encoding/json v1: duplicate keys are last-wins and invalid UTF-8 is coerced
// to U+FFFD. Dropping these two options switches the pipeline to v2's strict
// defaults (both cases rejected with an error).
var jsonv2Options = jsonv2.JoinOptions(
	jsontext.AllowDuplicateNames(true),
	jsontext.AllowInvalidUTF8(true),
)

type jsonv2Decoder struct{}

func (jsonv2Decoder) Unmarshal(buf []byte, val any) error {
	return jsonv2.Unmarshal(buf, val, jsonv2Options)
}

// jsonDecoder decodes request bodies with encoding/json/v2 (Go 1.27+).
// Strings are always copied (invariant I3), raw control characters are
// rejected and nesting is capped at jsontext's max depth of 10000.
var jsonDecoder jsonv2Decoder

func payloadMaxBytes() int64 {
	if n := settings.ServerSettings.PayloadMaxBytes; n != 0 {
		return n
	}
	return defaultPayloadMaxBytes
}

// bindJSONPayload reads the size-limited request body and decodes it into dst.
func bindJSONPayload(c *gin.Context, dst any) error {
	if c.Request == nil || c.Request.Body == nil {
		return errors.New("invalid request")
	}
	body := io.Reader(c.Request.Body)
	if limit := payloadMaxBytes(); limit >= 0 {
		body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
	}
	raw, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	return jsonDecoder.Unmarshal(raw, dst)
}
