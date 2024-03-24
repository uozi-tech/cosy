package cosy

import (
	"errors"
	"git.uozi.org/uozi/cosy/logger"
	"github.com/gin-gonic/gin"
	val "github.com/go-playground/validator/v10"
	"net/http"
	"reflect"
	"strings"
)

type ValidError struct {
	Key     string
	Message string
}

func (c *Ctx[T]) validate() (errs gin.H) {
	c.Payload = make(gin.H)

	_ = c.ShouldBindJSON(&c.Payload)

	//	logger.Debug(c.Payload, c.rules)

	errs = validate.ValidateMap(c.Payload, c.rules)

	if len(errs) > 0 {
		logger.Debug(errs)
		for k := range errs {
			errs[k] = c.rules[k]
		}
		return
	}
	// Make sure that the key in c.Payload is also the key of rules
	validated := make(map[string]interface{})

	for k, v := range c.Payload {
		if _, ok := c.rules[k]; ok {
			validated[k] = v
		}
	}

	c.Payload = validated

	return
}

func BindAndValid(c *gin.Context, target interface{}) bool {
	err := c.ShouldBindJSON(target)
	if err != nil {
		logger.Error("bind err", err)

		var verrs val.ValidationErrors
		ok := errors.As(err, &verrs)

		if !ok {
			c.JSON(http.StatusNotAcceptable, gin.H{
				"message": "Requested with wrong parameters",
				"code":    http.StatusNotAcceptable,
			})
			return false
		}

		t := reflect.TypeOf(target).Elem()
		errorsMap := make(map[string]interface{})
		for _, value := range verrs {
			var path []string

			namespace := strings.Split(value.StructNamespace(), ".")
			logger.Debug(t.Name(), namespace)
			if t.Name() != "" && len(namespace) > 1 {
				namespace = namespace[1:]
			}

			getJsonPath(t, namespace, &path)
			insertError(errorsMap, path, value.Tag())
		}

		c.JSON(http.StatusNotAcceptable, gin.H{
			"errors":  errorsMap,
			"message": "Requested with wrong parameters",
			"code":    http.StatusNotAcceptable,
		})

		return false
	}

	return true
}

// findField recursively finds the field in a nested struct
func getJsonPath(t reflect.Type, fields []string, path *[]string) {
	f, ok := t.FieldByName(fields[0])
	if !ok {
		return
	}

	*path = append(*path, f.Tag.Get("json"))

	if len(fields) > 1 {
		subFields := fields[1:]
		getJsonPath(f.Type, subFields, path)
	}
}

// insertError inserts an error into the errors map
func insertError(errorsMap map[string]interface{}, path []string, errorTag string) {
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
		errorsMap[jsonTag] = make(map[string]interface{})
	}

	// Recursively insert into the nested map
	subMap, _ := errorsMap[jsonTag].(map[string]interface{})
	insertError(subMap, path[1:], errorTag)
}
