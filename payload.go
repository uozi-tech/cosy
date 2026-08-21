package cosy

import (
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

// decodeJSON decodes a request body with encoding/json/v2 (Go 1.27+) using
// its strict defaults — a deliberate decision (PERF_REFACTOR_PLAN.md P3):
//
//   - duplicate object keys are rejected instead of last-wins, closing the
//     parser-differential ambiguity of corpus E;
//   - invalid UTF-8 is rejected instead of being coerced to U+FFFD (corpus D);
//   - raw control characters are rejected, strings are always copied (I3)
//     and nesting is capped at jsontext's max depth of 10000.
//
// Do not re-add jsontext.AllowDuplicateNames / AllowInvalidUTF8 for
// convenience; both loosen a security property.
func decodeJSON(raw []byte, dst any) error {
	return jsonv2.Unmarshal(raw, dst)
}

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
	return decodeJSON(raw, dst)
}
