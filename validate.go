package cosy

import (
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/uozi-tech/cosy/internal/rulecheck"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/valid"
)

type ValidateError struct {
	// CosyError
	Error
	// ValidateErrors
	Errors map[string]any `json:"errors"`
}

func NewValidateError(errors map[string]any) *ValidateError {
	return &ValidateError{
		Scope:   "validate",
		Code:    http.StatusNotAcceptable,
		Message: "Requested with wrong parameters",
		Errors:  errors,
	}
}

var v *validator.Validate

func init() {
	var ok bool
	v, ok = binding.Validator.Engine().(*validator.Validate)
	if !ok {
		logger.Fatal("failed to initialize binding validator engine")
	}

	err := v.RegisterValidation("date", valid.IsDate)
	if err != nil {
		logger.Fatal(err)
	}

	err = v.RegisterValidation("safety_text", valid.SafetyText)
	if err != nil {
		logger.Fatal(err)
	}
}

// GetValidator returns the validator instance.
//
// To override one of the rule names the compiled fast path implements
// (email, url, date, safety_text, hostname_port, min, max, oneof) use
// RegisterValidation: registering directly on the returned instance only
// affects rules that reach the validator fallback.
func GetValidator() *validator.Validate {
	return v
}

// RegisterValidation registers a custom validation on the shared validator and
// marks the tag as overridden so the compiled rule engine routes it to the
// validator instead of its built-in implementation.
func RegisterValidation(tag string, fn validator.Func, callValidationEvenIfNull ...bool) error {
	if err := v.RegisterValidation(tag, fn, callValidationEvenIfNull...); err != nil {
		return err
	}
	rulecheck.Override(tag)
	return nil
}

// ErrEmptyPayload is the "body" entry of the 406 ValidateError returned by
// Modify and BatchModify when, after rule filtering and hooks, the request
// carries no field the endpoint is allowed to update.
const ErrEmptyPayload = "empty payload"

// abortEmptyPayload rejects an update that would select no column. Letting it
// through would hand GORM an empty Select, which falls back to "*" and
// overwrites every column with the zero-valued model.
func abortEmptyPayload[T any](c *Ctx[T]) {
	c.JSON(http.StatusNotAcceptable, NewValidateError(gin.H{"body": ErrEmptyPayload}))
	c.Abort()
}

type ValidError struct {
	Key     string
	Message string
}

func (c *Ctx[T]) validate() (errs gin.H) {
	c.Payload = make(gin.H)

	if err := bindJSONPayload(c.Context, &c.Payload); err != nil {
		logJSONBindError(c.Context, err)
		return gin.H{"body": err.Error()}
	}
	if c.Payload == nil {
		c.Payload = make(gin.H)
	}

	// logger.Debug(c.Payload, c.rules)

	c.Payload["id"] = c.ID

	errs = rulecheck.ValidateMap(v, c.Payload, c.rules)

	if len(errs) > 0 {
		// logger.Debug(errs)
		for k := range errs {
			errs[k] = c.rules[k]
		}
		return
	}

	if len(c.unique) > 0 {
		conflicts, err := valid.DbUnique[T](c.Context, c.Payload, c.unique, c.columnMapping)
		if err != nil {
			c.AbortWithError(err)
			return
		}
		if len(conflicts) > 0 {
			// rulecheck.ValidateMap returns a nil map when everything passed
			if errs == nil {
				errs = make(gin.H, len(conflicts))
			}
			for _, v := range conflicts {
				errs[v] = "db_unique"
			}
			return
		}
	}

	// Make sure that the key in c.Payload is also the key of rules (I1):
	// drop everything else in place rather than rebuilding the map.
	for k := range c.Payload {
		if _, ok := c.rules[k]; !ok {
			delete(c.Payload, k)
		}
	}

	return
}

func validateBatchUpdate[T any](c *Ctx[T]) (errs gin.H) {
	c.Payload = make(gin.H)

	if err := bindJSONPayload(c.Context, &c.Payload); err != nil {
		logJSONBindError(c.Context, err)
		return gin.H{"body": err.Error()}
	}
	if c.Payload == nil {
		c.Payload = make(gin.H)
	}

	// logger.Debug(c.Payload, c.rules)

	if _, ok := c.Payload["ids"]; !ok {
		errs = gin.H{"ids": "required"}
		return
	}

	data, ok := c.Payload["data"].(map[string]any)
	if !ok {
		errs = gin.H{"data": "required"}
		return
	}

	errs = rulecheck.ValidateMap(v, data, c.rules)

	if len(errs) > 0 {
		// logger.Debug(errs)
		for k := range errs {
			errs[k] = c.rules[k]
		}
		return
	}

	// Make sure that the key in data is also the key of rules (I1)
	for k := range data {
		if _, ok := c.rules[k]; !ok {
			delete(data, k)
		}
	}
	c.Payload["data"] = data

	return
}

func logJSONBindError(c *gin.Context, err error) {
	logger.NewSessionLogger(c).Errorf("failed to bind JSON request body: %v", err)
}

func BindAndValid(c *gin.Context, target any) bool {
	if err := bindJSONPayload(c, target); err != nil {
		return abortBindError(c, err)
	}
	if binding.Validator == nil {
		return true
	}

	// Allocated on the first failure only, a valid request pays nothing for it
	var errorsMap map[string]any
	if err := collectValidationErrors(reflect.ValueOf(target), nil, &errorsMap); err != nil {
		errHandler(c, err)
		return false
	}
	if len(errorsMap) > 0 {
		c.JSON(http.StatusNotAcceptable, NewValidateError(errorsMap))
		return false
	}

	return true
}

// collectValidationErrors validates value and records every failure in
// errorsMap under its JSON path. Slices and arrays are walked here rather than
// handed to gin, whose SliceValidationError drops the index of the failing
// element. An error is returned only for failures that are not validation
// errors.
func collectValidationErrors(value reflect.Value, prefix []string, errorsMap *map[string]any) error {
	for value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			// Copy the prefix so sibling elements never share a backing array
			elemPrefix := append(append([]string(nil), prefix...), strconv.Itoa(i))
			if err := collectValidationErrors(value.Index(i), elemPrefix, errorsMap); err != nil {
				return err
			}
		}
		return nil
	case reflect.Struct:
	default:
		return nil
	}

	var obj any
	if value.CanAddr() {
		obj = value.Addr().Interface()
	} else {
		obj = value.Interface()
	}

	err := binding.Validator.ValidateStruct(obj)
	if err == nil {
		return nil
	}
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return err
	}

	if *errorsMap == nil {
		*errorsMap = make(map[string]any)
	}

	t := value.Type()
	for _, fieldErr := range verrs {
		namespace := fieldErr.StructNamespace()
		// The namespace starts with the name of the validated struct, unless
		// the struct is anonymous
		if name := t.Name(); name != "" {
			namespace = strings.TrimPrefix(namespace, name+".")
		}

		path := append(append([]string(nil), prefix...), getJsonPath(t, namespace)...)
		insertError(*errorsMap, path, fieldErr.Tag())
	}
	return nil
}

// abortBindError answers a failed body read/decode: oversized bodies get 413,
// malformed or ill-typed JSON gets 406 (same shape as the CRUD pipeline), and
// anything else is a server-side failure handled by errHandler.
func abortBindError(c *gin.Context, err error) bool {
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &tooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, &ValidateError{
			Scope:   "validate",
			Code:    http.StatusRequestEntityTooLarge,
			Message: "Request body too large",
			Errors:  gin.H{"body": err.Error()},
		})
	case isPayloadError(err):
		c.JSON(http.StatusNotAcceptable, NewValidateError(gin.H{"body": err.Error()}))
	default:
		errHandler(c, err)
	}
	return false
}

// namespaceSegment is one dot separated part of a validator namespace: a
// struct field name followed by the slice indexes or map keys applied to it,
// e.g. "Items[0][key]".
type namespaceSegment struct {
	name string
	keys []string
}

// splitNamespace splits a validator namespace such as "Detail.Items[0].Name"
// into its segments. A map key may contain dots, so a bracket is only closed by
// a "]" that ends the segment or is followed by another key.
func splitNamespace(namespace string) []namespaceSegment {
	var segments []namespaceSegment

	for len(namespace) > 0 {
		var segment namespaceSegment

		end := strings.IndexAny(namespace, ".[")
		if end < 0 {
			end = len(namespace)
		}
		segment.name = namespace[:end]
		namespace = namespace[end:]

		for strings.HasPrefix(namespace, "[") {
			closing := -1
			for i := 1; i < len(namespace); i++ {
				if namespace[i] == ']' &&
					(i+1 == len(namespace) || namespace[i+1] == '.' || namespace[i+1] == '[') {
					closing = i
					break
				}
			}
			if closing < 0 {
				// Unbalanced bracket: keep the rest as the key instead of dropping it
				segment.keys = append(segment.keys, namespace[1:])
				namespace = ""
				break
			}
			segment.keys = append(segment.keys, namespace[1:closing])
			namespace = namespace[closing+1:]
		}

		segments = append(segments, segment)
		namespace = strings.TrimPrefix(namespace, ".")
	}

	return segments
}

// indirectType strips every pointer level from t
func indirectType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// jsonFieldName returns the name encoding/json uses for the field. flatten
// reports an embedded struct without a json name, whose fields are promoted to
// the parent object and therefore add nothing to the path.
func jsonFieldName(f reflect.StructField) (name string, flatten bool) {
	name, _, _ = strings.Cut(f.Tag.Get("json"), ",")
	if name != "" && name != "-" {
		return name, false
	}
	if f.Anonymous && name == "" && indirectType(f.Type).Kind() == reflect.Struct {
		return "", true
	}
	// Use lowercase field name as fallback
	return strings.ToLower(f.Name), false
}

// getJsonPath converts a validator struct namespace, relative to t, into the
// path of the offending value in the JSON document. Pointers, slices, arrays
// and maps are followed down to their element type. Whatever cannot be
// resolved from the static type (a struct behind an interface, an unknown
// field) falls back to the lowercase field name, so the path is never
// truncated and resolving it never panics.
func getJsonPath(t reflect.Type, namespace string) []string {
	var path []string

	segments := splitNamespace(namespace)
	for _, segment := range segments {
		name := strings.ToLower(segment.name)

		t = indirectType(t)
		if t != nil && t.Kind() == reflect.Struct {
			if f, ok := t.FieldByName(segment.name); ok {
				// name is empty for a flattened embedded struct
				name, _ = jsonFieldName(f)
				t = f.Type
			} else {
				t = nil
			}
		} else {
			t = nil
		}

		if name != "" {
			path = append(path, name)
		}
		path = append(path, segment.keys...)

		for range segment.keys {
			t = indirectType(t)
			if t == nil {
				break
			}
			switch t.Kind() {
			case reflect.Slice, reflect.Array, reflect.Map:
				t = t.Elem()
			default:
				t = nil
			}
		}
	}

	// The failing field is a flattened embedded struct itself (e.g. a required
	// embedded pointer): report it under its own name rather than losing it
	if len(path) == 0 && len(segments) > 0 {
		path = append(path, strings.ToLower(segments[len(segments)-1].name))
	}

	return path
}

// insertError inserts an error into the errors map. The first error recorded
// at a path wins: a later one never replaces it nor descends through it.
func insertError(errorsMap map[string]any, path []string, errorTag string) {
	if len(path) == 0 {
		return
	}

	jsonTag := path[0]
	if len(path) == 1 {
		// Last element in the path, set the error
		if _, ok := errorsMap[jsonTag]; !ok {
			errorsMap[jsonTag] = errorTag
		}
		return
	}

	// Create a new map if necessary
	if _, ok := errorsMap[jsonTag]; !ok {
		errorsMap[jsonTag] = make(map[string]any)
	}

	// Recursively insert into the nested map
	subMap, ok := errorsMap[jsonTag].(map[string]any)
	if !ok {
		return
	}
	insertError(subMap, path[1:], errorTag)
}
