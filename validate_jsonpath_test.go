package cosy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jsonPathDetail struct {
	Year  int    `json:"year" binding:"omitempty,min=1900"`
	Plate string `json:"plate" binding:"omitempty,max=3"`
}

type jsonPathEmbedded struct {
	Code string `json:"code" binding:"omitempty,max=3"`
}

type jsonPathTagged struct {
	Label string `json:"label" binding:"omitempty,max=3"`
}

type jsonPathRequest struct {
	jsonPathEmbedded
	jsonPathTagged `json:"tagged"`

	Title       string                     `json:"title" binding:"omitempty,max=3"`
	Detail      *jsonPathDetail            `json:"detail"`
	DeepDetail  **jsonPathDetail           `json:"deep_detail"`
	Items       []jsonPathDetail           `json:"items" binding:"dive"`
	ItemPtrs    []*jsonPathDetail          `json:"item_ptrs" binding:"dive"`
	ItemsPtr    *[]jsonPathDetail          `json:"items_ptr" binding:"omitempty,dive"`
	Array       [1]jsonPathDetail          `json:"array" binding:"dive"`
	Matrix      [][]jsonPathDetail         `json:"matrix" binding:"dive,dive"`
	ByKey       map[string]jsonPathDetail  `json:"by_key" binding:"dive"`
	ByKeyPtr    map[string]*jsonPathDetail `json:"by_key_ptr" binding:"dive"`
	ByID        map[int][]jsonPathDetail   `json:"by_id" binding:"dive,dive"`
	Names       map[string]string          `json:"names" binding:"dive,max=3"`
	NoTag       *jsonPathDetail
	OmitNameTag *jsonPathDetail `json:",omitempty"`
}

func serveBindAndValid[T any](t *testing.T, body string) (int, map[string]any) {
	t.Helper()

	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		var data T
		if !BindAndValid(c, &data) {
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	// Without a recovery middleware a panic inside BindAndValid would take the
	// whole test binary down, so convert it into a failure of this case only.
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("BindAndValid panicked: %v", recovered)
			}
		}()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body)))
	}()

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp), w.Body.String())
	return w.Code, resp
}

// TestBindAndValidJsonPath covers the error path building for every field
// shape the validator can descend into. The pointer cases reproduce
// "reflect: FieldByName of non-struct type *model.InsuranceDetailFields".
func TestBindAndValidJsonPath(t *testing.T) {
	cases := []struct {
		name string
		body string
		want map[string]any
	}{
		{
			name: "plain field",
			body: `{"title":"toolong"}`,
			want: map[string]any{"title": "max"},
		},
		{
			name: "pointer to struct",
			body: `{"detail":{"year":1800,"plate":"toolong"}}`,
			want: map[string]any{"detail": map[string]any{"year": "min", "plate": "max"}},
		},
		{
			name: "pointer to pointer to struct",
			body: `{"deep_detail":{"year":1800}}`,
			want: map[string]any{"deep_detail": map[string]any{"year": "min"}},
		},
		{
			name: "slice of struct",
			body: `{"items":[{"year":2000},{"year":1800}]}`,
			want: map[string]any{"items": map[string]any{"1": map[string]any{"year": "min"}}},
		},
		{
			name: "slice of pointer to struct",
			body: `{"item_ptrs":[{"plate":"toolong"}]}`,
			want: map[string]any{"item_ptrs": map[string]any{"0": map[string]any{"plate": "max"}}},
		},
		{
			name: "pointer to slice of struct",
			body: `{"items_ptr":[{"year":1800}]}`,
			want: map[string]any{"items_ptr": map[string]any{"0": map[string]any{"year": "min"}}},
		},
		{
			name: "array of struct",
			body: `{"array":[{"year":1800}]}`,
			want: map[string]any{"array": map[string]any{"0": map[string]any{"year": "min"}}},
		},
		{
			name: "nested slices",
			body: `{"matrix":[[{"year":2000}],[{"year":2000},{"year":1800}]]}`,
			want: map[string]any{"matrix": map[string]any{"1": map[string]any{"1": map[string]any{"year": "min"}}}},
		},
		{
			name: "map of struct",
			body: `{"by_key":{"car-1":{"year":1800}}}`,
			want: map[string]any{"by_key": map[string]any{"car-1": map[string]any{"year": "min"}}},
		},
		{
			name: "map of pointer to struct",
			body: `{"by_key_ptr":{"a.b":{"year":1800}}}`,
			want: map[string]any{"by_key_ptr": map[string]any{"a.b": map[string]any{"year": "min"}}},
		},
		{
			name: "map of slice of struct",
			body: `{"by_id":{"7":[{"year":1800}]}}`,
			want: map[string]any{"by_id": map[string]any{"7": map[string]any{"0": map[string]any{"year": "min"}}}},
		},
		{
			name: "map of scalar",
			body: `{"names":{"first":"toolong"}}`,
			want: map[string]any{"names": map[string]any{"first": "max"}},
		},
		{
			name: "embedded struct is flattened like encoding/json",
			body: `{"code":"toolong"}`,
			want: map[string]any{"code": "max"},
		},
		{
			name: "embedded struct with json tag is nested",
			body: `{"tagged":{"label":"toolong"}}`,
			want: map[string]any{"tagged": map[string]any{"label": "max"}},
		},
		{
			name: "field without json tag",
			body: `{"NoTag":{"year":1800}}`,
			want: map[string]any{"notag": map[string]any{"year": "min"}},
		},
		{
			name: "json tag without a name",
			body: `{"OmitNameTag":{"year":1800}}`,
			want: map[string]any{"omitnametag": map[string]any{"year": "min"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, resp := serveBindAndValid[jsonPathRequest](t, tc.body)
			assert.Equal(t, http.StatusNotAcceptable, code)
			assert.Equal(t, tc.want, resp["errors"])
		})
	}
}

func TestBindAndValidJsonPathPassesValidPayload(t *testing.T) {
	code, _ := serveBindAndValid[jsonPathRequest](t,
		`{"title":"ok","detail":{"year":2000},"items":[{"year":2000}],"by_key":{"a":{"plate":"abc"}}}`)
	assert.Equal(t, http.StatusOK, code)
}

// A top level slice is validated element by element by gin, which reports the
// failures as binding.SliceValidationError rather than ValidationErrors.
func TestBindAndValidJsonPathTopLevelSlice(t *testing.T) {
	code, resp := serveBindAndValid[[]jsonPathDetail](t, `[{"year":2000},{"year":1800}]`)
	assert.Equal(t, http.StatusNotAcceptable, code)
	assert.Equal(t, map[string]any{"1": map[string]any{"year": "min"}}, resp["errors"])
}

func TestGetJsonPathNeverPanics(t *testing.T) {
	typ := reflect.TypeOf(jsonPathRequest{})
	for namespace, want := range map[string][]string{
		"":                           nil,
		"Missing":                    {"missing"},
		"Title.Missing":              {"title", "missing"},
		"Detail.Missing":             {"detail", "missing"},
		"Items[0].Missing.Deeper":    {"items", "0", "missing", "deeper"},
		"Names[first].Missing":       {"names", "first", "missing"},
		"Title[0][1].Year":           {"title", "0", "1", "year"},
		"Items[.Year":                {"items", ".Year"},
		".Year":                      {"year"},
		"ByKeyPtr[a.b][c].Year":      {"by_key_ptr", "a.b", "c", "year"},
		"jsonPathEmbedded":           {"jsonpathembedded"},
		"jsonPathEmbedded.Code":      {"code"},
		"jsonPathTagged.Label":       {"tagged", "label"},
		"Matrix[1][2].Plate":         {"matrix", "1", "2", "plate"},
		"ByID[7][0].Year":            {"by_id", "7", "0", "year"},
		"DeepDetail.Plate":           {"deep_detail", "plate"},
		"OmitNameTag.Year":           {"omitnametag", "year"},
		"Detail.Year.Deeper[0].Leaf": {"detail", "year", "deeper", "0", "leaf"},
	} {
		assert.NotPanics(t, func() {
			assert.Equal(t, want, getJsonPath(typ, namespace), namespace)
		}, namespace)
	}
}

// A struct hidden behind an interface cannot be resolved from the static type,
// the path falls back to the lowercase field names instead of being truncated.
func TestBindAndValidJsonPathInterfaceField(t *testing.T) {
	type request struct {
		Payload any `json:"payload"`
	}

	var errorsMap map[string]any
	req := request{Payload: &jsonPathDetail{Year: 1800}}
	require.NoError(t, collectValidationErrors(reflect.ValueOf(&req), nil, &errorsMap))
	assert.Equal(t, map[string]any{"payload": map[string]any{"year": "min"}}, errorsMap)
}

func TestInsertErrorKeepsFirstErrorOnCollision(t *testing.T) {
	errorsMap := make(map[string]any)

	assert.NotPanics(t, func() {
		insertError(errorsMap, []string{"items"}, "min")
		insertError(errorsMap, []string{"items", "0", "year"}, "max")
		insertError(errorsMap, []string{"detail", "year"}, "min")
		insertError(errorsMap, []string{"detail"}, "required")
		insertError(errorsMap, nil, "ignored")
	})

	assert.Equal(t, map[string]any{
		"items":  "min",
		"detail": map[string]any{"year": "min"},
	}, errorsMap)
}
