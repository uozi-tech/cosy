package cosy

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uozi-tech/cosy/settings"
)

type p3PayloadModel struct {
	Name string `json:"name"`
	Bio  string `json:"bio"`
}

func newP3ValidateCtx(t *testing.T, body string) *Ctx[p3PayloadModel] {
	t.Helper()
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodPost, "/models", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	return Core[p3PayloadModel](context).SetValidRules(gin.H{"name": "required"})
}

func withPayloadMaxBytes(t *testing.T, n int64) {
	t.Helper()
	prev := settings.ServerSettings.PayloadMaxBytes
	settings.ServerSettings.PayloadMaxBytes = n
	t.Cleanup(func() { settings.ServerSettings.PayloadMaxBytes = prev })
}

// F2: oversized bodies must be rejected before parsing.
func TestValidateRejectsOversizedBody(t *testing.T) {
	withPayloadMaxBytes(t, 64)

	body := `{"name":"` + strings.Repeat("a", 128) + `"}`
	errs := newP3ValidateCtx(t, body).validate()

	require.Contains(t, errs, "body")
	assert.Contains(t, errs["body"], "request body too large")
}

func TestValidateBodyLimitDisabled(t *testing.T) {
	withPayloadMaxBytes(t, -1)

	body := `{"name":"` + strings.Repeat("a", 1<<10) + `"}`
	errs := newP3ValidateCtx(t, body).validate()
	assert.Empty(t, errs)
}

func TestValidateBatchUpdateRejectsOversizedBody(t *testing.T) {
	withPayloadMaxBytes(t, 64)

	body := `{"ids":[1],"data":{"name":"` + strings.Repeat("a", 128) + `"}}`
	errs := validateBatchUpdate(newP3ValidateCtx(t, body))

	require.Contains(t, errs, "body")
	assert.Contains(t, errs["body"], "request body too large")
}

// Corpus C (DoS): deeply nested JSON must fail with an error, never panic or
// recurse unbounded. sonic caps nesting at depth 4096.
func TestValidateDeepNestingReturnsError(t *testing.T) {
	depth := 100_000
	body := `{"a":` + strings.Repeat("[", depth) + strings.Repeat("]", depth) + `}`
	errs := newP3ValidateCtx(t, body).validate()

	require.Contains(t, errs, "body")
	assert.Contains(t, errs["body"], "max depth")
}

// Corpus C: a wide-but-legal object inside the size limit still decodes.
func TestValidateWideMapWithinLimitSucceeds(t *testing.T) {
	var sb strings.Builder
	sb.WriteString(`{"name":"ok"`)
	for i := range 10_000 {
		sb.WriteString(`,"k`)
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(`":1`)
	}
	sb.WriteString(`}`)

	errs := newP3ValidateCtx(t, sb.String()).validate()
	assert.Empty(t, errs)
}

func TestValidateEmptyBodyReturnsError(t *testing.T) {
	errs := newP3ValidateCtx(t, "").validate()
	require.Contains(t, errs, "body")
}

// I3: decoded strings must never alias the request buffer (CopyString=true).
func TestJSONDecoderCopiesStrings(t *testing.T) {
	raw := []byte(`{"name":"copystring-check"}`)
	var m map[string]any
	require.NoError(t, jsonDecoder.Unmarshal(raw, &m))

	for i := range raw {
		raw[i] = 'z'
	}
	assert.Equal(t, "copystring-check", m["name"])
}

// Parity with encoding/json: raw control characters inside string values are
// rejected (ValidateString=true).
func TestJSONDecoderRejectsRawControlChars(t *testing.T) {
	var m map[string]any
	err := jsonDecoder.Unmarshal([]byte("{\"a\":\"x\x01y\"}"), &m)
	assert.Error(t, err)
}

type p3BindModel struct {
	Name string `json:"name" binding:"required"`
}

// BindAndValid keeps running struct validation after the sonic switch.
func TestBindAndValidStillValidatesStruct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/models", strings.NewReader(`{}`))
	context.Request.Header.Set("Content-Type", "application/json")

	var target p3BindModel
	require.False(t, BindAndValid(context, &target))
	assert.Equal(t, http.StatusNotAcceptable, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"name":"required"`)

	context, _ = gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodPost, "/models", strings.NewReader(`{"name":"jack"}`))
	context.Request.Header.Set("Content-Type", "application/json")

	target = p3BindModel{}
	require.True(t, BindAndValid(context, &target))
	assert.Equal(t, "jack", target.Name)
}

func TestBindAndValidRejectsOversizedBody(t *testing.T) {
	withPayloadMaxBytes(t, 64)
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/models",
		strings.NewReader(`{"name":"`+strings.Repeat("a", 128)+`"}`))
	context.Request.Header.Set("Content-Type", "application/json")

	var target p3BindModel
	assert.False(t, BindAndValid(context, &target))
}
