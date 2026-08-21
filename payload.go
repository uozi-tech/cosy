package cosy

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/settings"
)

// defaultPayloadMaxBytes caps JSON request bodies parsed by the CRUD pipeline
// and BindAndValid when settings.ServerSettings.PayloadMaxBytes is 0 (F2).
const defaultPayloadMaxBytes int64 = 10 << 20 // 10 MiB

// jsonDecoder is selected per toolchain by build tags:
//
//   - payload_jsonv2.go (go1.27+): encoding/json/v2 — standard library and the
//     fastest option on Go 1.27, where sonic only offers its compat path
//     (slower than encoding/json itself).
//   - payload_sonic.go (<go1.27): sonic native JIT.
//
// Both paths keep the same semantics: decoded strings are copies (I3), raw
// control characters are rejected, duplicate keys are last-wins and invalid
// UTF-8 is coerced to U+FFFD.

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
