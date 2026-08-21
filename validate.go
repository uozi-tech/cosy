package cosy

import (
	"errors"
	"net/http"
	"reflect"
	"regexp"
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
		Error: Error{
			Scope:   "validate",
			Code:    http.StatusNotAcceptable,
			Message: "Requested with wrong parameters",
		},
		Errors: errors,
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
	if err := binding.Validator.ValidateStruct(target); err != nil {
		var verrs validator.ValidationErrors
		ok := errors.As(err, &verrs)
		if !ok {
			errHandler(c, err)
			return false
		}

		t := reflect.TypeOf(target).Elem()
		errorsMap := make(map[string]any)
		for _, value := range verrs {
			var path []string
			namespace := strings.Split(value.StructNamespace(), ".")
			// logger.Debug(t.Name(), namespace)
			if t.Name() != "" && len(namespace) > 1 {
				namespace = namespace[1:]
			}

			getJsonPath(t, namespace, &path)
			insertError(errorsMap, path, value.Tag())
		}

		c.JSON(http.StatusNotAcceptable, NewValidateError(errorsMap))

		return false
	}

	return true
}

// abortBindError answers a failed body read/decode: oversized bodies get 413,
// malformed or ill-typed JSON gets 406 (same shape as the CRUD pipeline), and
// anything else is a server-side failure handled by errHandler.
func abortBindError(c *gin.Context, err error) bool {
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &tooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, &ValidateError{
			Error: Error{
				Scope:   "validate",
				Code:    http.StatusRequestEntityTooLarge,
				Message: "Request body too large",
			},
			Errors: gin.H{"body": err.Error()},
		})
	case isPayloadError(err):
		c.JSON(http.StatusNotAcceptable, NewValidateError(gin.H{"body": err.Error()}))
	default:
		errHandler(c, err)
	}
	return false
}

// findField recursively finds the field in a nested struct
func getJsonPath(t reflect.Type, fields []string, path *[]string) {
	field := fields[0]
	// used in case of an array
	var index string
	if field[len(field)-1] == ']' {
		re := regexp.MustCompile(`(\w+)\[(\d+)\]`)
		matches := re.FindStringSubmatch(field)

		if len(matches) > 2 {
			field = matches[1]
			index = matches[2]
		}
	}

	f, ok := t.FieldByName(field)
	if !ok {
		return
	}

	jsonTag := f.Tag.Get("json")
	// Handle empty json tag case
	if jsonTag == "" {
		// Use lowercase field name as fallback
		jsonTag = strings.ToLower(field)
	} else {
		// Handle json:"name,omitempty" case
		if parts := strings.Split(jsonTag, ","); len(parts) > 0 {
			jsonTag = parts[0]
		}
	}

	*path = append(*path, jsonTag)

	if index != "" {
		*path = append(*path, index)
	}

	if len(fields) > 1 {
		subFields := fields[1:]
		getJsonPath(f.Type, subFields, path)
	}
}

// insertError inserts an error into the errors map
func insertError(errorsMap map[string]any, path []string, errorTag string) {
	if len(path) == 0 {
		return
	}

	jsonTag := path[0]
	if len(path) == 1 {
		// Last element in the path, set the error
		errorsMap[jsonTag] = errorTag
		return
	}

	// Create a new map if necessary
	if _, ok := errorsMap[jsonTag]; !ok {
		errorsMap[jsonTag] = make(map[string]any)
	}

	// Recursively insert into the nested map
	subMap, _ := errorsMap[jsonTag].(map[string]any)
	insertError(subMap, path[1:], errorTag)
}
