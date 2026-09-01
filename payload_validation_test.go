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

type validationPayload struct {
	Name string `json:"name"`
	Bio  string `json:"bio"`
}

func newJSONTestContext(
	t *testing.T,
	method, path, body string,
) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	return context, recorder
}

func newPayloadValidationContext(t *testing.T, body string) *Ctx[validationPayload] {
	t.Helper()
	context, _ := newJSONTestContext(t, http.MethodPost, "/models", body)
	return Core[validationPayload](context).SetValidRules(gin.H{"name": "required"})
}

func withPayloadMaxBytes(t *testing.T, n int64) {
	t.Helper()
	prev := settings.ServerSettings.PayloadMaxBytes
	settings.ServerSettings.PayloadMaxBytes = n
	t.Cleanup(func() { settings.ServerSettings.PayloadMaxBytes = prev })
}

// Oversized bodies must be rejected before parsing.
func TestValidateRejectsOversizedBody(t *testing.T) {
	withPayloadMaxBytes(t, 64)

	body := `{"name":"` + strings.Repeat("a", 128) + `"}`
	errs := newPayloadValidationContext(t, body).validate()

	require.Contains(t, errs, "body")
	assert.Contains(t, errs["body"], "request body too large")
}

func TestValidateBodyLimitDisabled(t *testing.T) {
	withPayloadMaxBytes(t, -1)

	body := `{"name":"` + strings.Repeat("a", 1<<10) + `"}`
	errs := newPayloadValidationContext(t, body).validate()
	assert.Empty(t, errs)
}

func TestValidateBatchUpdateRejectsOversizedBody(t *testing.T) {
	withPayloadMaxBytes(t, 64)

	body := `{"ids":[1],"data":{"name":"` + strings.Repeat("a", 128) + `"}}`
	errs := validateBatchUpdate(newPayloadValidationContext(t, body))

	require.Contains(t, errs, "body")
	assert.Contains(t, errs["body"], "request body too large")
}

// Deeply nested JSON must fail with an error instead of recursing unbounded.
func TestValidateDeepNestingReturnsError(t *testing.T) {
	depth := 100_000
	body := `{"a":` + strings.Repeat("[", depth) + strings.Repeat("]", depth) + `}`
	errs := newPayloadValidationContext(t, body).validate()

	require.Contains(t, errs, "body")
	assert.Contains(t, errs["body"], "max depth")
}

// A wide-but-legal object inside the size limit still decodes.
func TestValidateWideMapWithinLimitSucceeds(t *testing.T) {
	var sb strings.Builder
	sb.WriteString(`{"name":"ok"`)
	for i := range 10_000 {
		sb.WriteString(`,"k`)
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(`":1`)
	}
	sb.WriteString(`}`)

	errs := newPayloadValidationContext(t, sb.String()).validate()
	assert.Empty(t, errs)
}

func TestValidateEmptyBodyReturnsError(t *testing.T) {
	errs := newPayloadValidationContext(t, "").validate()
	require.Contains(t, errs, "body")
}

func TestValidateRejectsStrictJSONViolations(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		message string
	}{
		{name: "duplicate keys", body: `{"name":"user","name":"admin"}`, message: "duplicate"},
		{name: "invalid UTF-8", body: "{\"name\":\"\xff\"}", message: "invalid UTF-8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := newPayloadValidationContext(t, tt.body).validate()
			require.Contains(t, errs, "body")
			assert.Contains(t, errs["body"], tt.message)
		})
	}
}

type bindingPayload struct {
	Name string `json:"name" binding:"required"`
}

// BindAndValid decodes the payload and applies struct validation.
func TestBindAndValidValidatesStruct(t *testing.T) {
	context, recorder := newJSONTestContext(t, http.MethodPost, "/models", `{}`)

	var target bindingPayload
	require.False(t, BindAndValid(context, &target))
	assert.Equal(t, http.StatusNotAcceptable, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"name":"required"`)

	context, _ = newJSONTestContext(t, http.MethodPost, "/models", `{"name":"jack"}`)

	target = bindingPayload{}
	require.True(t, BindAndValid(context, &target))
	assert.Equal(t, "jack", target.Name)
}

func TestBindAndValidRejectsOversizedBody(t *testing.T) {
	withPayloadMaxBytes(t, 64)
	context, recorder := newJSONTestContext(t, http.MethodPost, "/models",
		`{"name":"`+strings.Repeat("a", 128)+`"}`)

	var target bindingPayload
	assert.False(t, BindAndValid(context, &target))
	assert.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
}

// encoding/json v1 matched member names case-insensitively; BindAndValid keeps
// that contract on json/v2.
func TestBindAndValidMatchesNamesCaseInsensitively(t *testing.T) {
	context, _ := newJSONTestContext(t, http.MethodPost, "/models", `{"Name":"jack"}`)
	var target bindingPayload
	require.True(t, BindAndValid(context, &target))
	assert.Equal(t, "jack", target.Name)
}

// Strict decoding failures are client faults: 406, never 500.
func TestBindAndValidAnswers406ForDuplicateKeys(t *testing.T) {
	context, recorder := newJSONTestContext(t, http.MethodPost, "/models", `{"name":"a","name":"b"}`)
	var target bindingPayload
	require.False(t, BindAndValid(context, &target))
	assert.Equal(t, http.StatusNotAcceptable, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "duplicate")
}

func TestValidateBatchUpdateRejectsNonObjectData(t *testing.T) {
	for _, body := range []string{`{"ids":[1],"data":[]}`, `{"ids":[1],"data":null}`, `{"ids":[1],"data":1}`} {
		errs := validateBatchUpdate(newPayloadValidationContext(t, body))
		assert.Equal(t, gin.H{"data": "required"}, errs, body)
	}
}
