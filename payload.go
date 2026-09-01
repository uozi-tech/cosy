package cosy

import (
	jsonv1 "encoding/json"
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/settings"
)

// decodeOptions defines the contract shared by every request-body decode:
//
//   - DefaultOptionsV1 keeps encoding/json v1 semantics for struct targets
//     (case-insensitive member matching, null leaves pre-populated fields,
//     `,string` and other legacy tag shapes keep working), so BindAndValid
//     behaves as it did under gin's ShouldBindJSON;
//   - on top of that, duplicate object keys are rejected instead of
//     last-wins (closes the parser-differential ambiguity of corpus E) and
//     invalid UTF-8 is rejected instead of being coerced to U+FFFD (corpus D).
//
// Raw control characters are rejected, strings are always copied (I3) and
// nesting is capped at the decoder's max depth. Do not re-add
// jsontext.AllowDuplicateNames / AllowInvalidUTF8 for convenience; both
// loosen a security property.
var decodeOptions = jsonv2.JoinOptions(
	jsonv1.DefaultOptionsV1(),
	jsontext.AllowDuplicateNames(false),
	jsontext.AllowInvalidUTF8(false),
)

// decodeJSON decodes a request body with encoding/json/v2 and decodeOptions.
func decodeJSON(raw []byte, dst any) error {
	return jsonv2.Unmarshal(raw, dst, decodeOptions)
}

// isPayloadError reports whether err describes a malformed or ill-typed
// request body (as opposed to an I/O or server-side failure).
func isPayloadError(err error) bool {
	var (
		syntactic *jsontext.SyntacticError
		semantic  *jsonv2.SemanticError
		v1Syntax  *jsonv1.SyntaxError
		v1Type    *jsonv1.UnmarshalTypeError
	)
	return errors.As(err, &syntactic) || errors.As(err, &semantic) ||
		errors.As(err, &v1Syntax) || errors.As(err, &v1Type) ||
		errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF)
}

func payloadMaxBytes() int64 {
	return settings.ServerSettings.PayloadLimit()
}

// bindJSONPayload reads the size-limited request body and decodes it into dst.
func bindJSONPayload(c *gin.Context, dst any) error {
	if c.Request == nil || c.Request.Body == nil {
		return errors.New("invalid request")
	}
	body := io.Reader(c.Request.Body)
	limit := payloadMaxBytes()
	if limit >= 0 {
		body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
	}
	raw, err := readBody(body, c.Request.ContentLength, limit)
	if err != nil {
		return err
	}
	return decodeJSON(raw, dst)
}

// readBody is io.ReadAll with the buffer sized up front from Content-Length
// (capped by the body limit), so a typical request is read in one allocation
// instead of doubling from 512 bytes.
func readBody(body io.Reader, contentLength, limit int64) ([]byte, error) {
	size := int64(512)
	if contentLength > 0 {
		size = contentLength + 1 // +1 so the final Read sees io.EOF without growing
	}
	if limit >= 0 && size > limit+1 {
		size = limit + 1
	}
	buf := make([]byte, 0, size)
	for {
		if len(buf) == cap(buf) {
			buf = append(buf, 0)[:len(buf)]
		}
		n, err := body.Read(buf[len(buf):cap(buf)])
		buf = buf[:len(buf)+n]
		if err != nil {
			if err == io.EOF {
				return buf, nil
			}
			return buf, err
		}
	}
}
