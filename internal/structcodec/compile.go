package structcodec

import (
	"fmt"
	"reflect"
	"strings"
)

func compilePlan(typ reflect.Type) (*decodePlan, error) {
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("structcodec: cannot compile %s", typ)
	}

	plan := &decodePlan{typ: typ}
	walkFields(typ, nil, &plan.fields)
	return plan, nil
}

func walkFields(typ reflect.Type, prefix []addressStep, fields *[]fieldPlan) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("json")
		parts := strings.Split(tag, ",")
		name := field.Name
		if len(parts) > 0 && parts[0] != "" {
			name = parts[0]
		}

		squash := field.Anonymous
		for _, option := range parts[1:] {
			if option == "squash" {
				squash = true
			}
		}

		steps := appendSteps(prefix, addressStep{offset: field.Offset})
		fieldType := field.Type
		if squash && fieldType.Kind() == reflect.Struct {
			walkFields(fieldType, steps, fields)
			continue
		}
		if squash && fieldType.Kind() == reflect.Pointer && fieldType.Elem().Kind() == reflect.Struct {
			steps[len(steps)-1].ptrType = fieldType
			walkFields(fieldType.Elem(), steps, fields)
			continue
		}

		if field.PkgPath != "" {
			continue
		}
		*fields = append(*fields, fieldPlan{
			name:   name,
			steps:  steps,
			decode: compileValueDecoder(fieldType),
		})
	}
}

func appendSteps(prefix []addressStep, step addressStep) []addressStep {
	steps := make([]addressStep, len(prefix)+1)
	copy(steps, prefix)
	steps[len(prefix)] = step
	return steps
}
