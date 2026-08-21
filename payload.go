package cosy

import (
	"errors"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/settings"
)

// defaultPayloadMaxBytes caps JSON request bodies parsed by the CRUD pipeline
// and BindAndValid when settings.ServerSettings.PayloadMaxBytes is 0 (F2).
const defaultPayloadMaxBytes int64 = 10 << 20 // 10 MiB

// jsonDecoder is the frozen sonic configuration for decoding request bodies.
//
// Config discipline (PERF_REFACTOR_PLAN.md §3.3, invariant I3): CopyString
// must stay true so decoded strings never alias the request buffer, and
// ValidateString must stay true so raw control characters in string values
// are rejected as encoding/json does. Do not switch to ConfigDefault or
// ConfigFastest for speed — ConfigDefault has CopyString=false.
var jsonDecoder = sonic.Config{
	CopyString:     true,
	ValidateString: true,
}.Froze()

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
